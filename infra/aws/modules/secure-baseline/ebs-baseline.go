package secure_baseline

import (
	"fmt"

	"github.com/pulumi/pulumi-aws/sdk/v4/go/aws/ebs"
	"github.com/pulumi/pulumi-aws/sdk/v4/go/aws/iam"
	"github.com/pulumi/pulumi-aws/sdk/v4/go/aws/kms"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func EbsBaseline(ctx *pulumi.Context, name string, accountId string, policy *iam.GetPolicyDocumentResult) (*kms.Key, error) {

	key, err := kms.NewKey(ctx, "ebs", &kms.KeyArgs{
		Description: pulumi.String("KMS key ebs baseline"),
		Policy:      pulumi.String(policy.Json),
	})
	if err != nil {
		return nil, err
	}
	_, err = kms.NewAlias(ctx, "ebs-encryption-alias", &kms.AliasArgs{
		TargetKeyId: key.KeyId,
		Name: pulumi.String("alias/ebs-encryption"),
	})
	if err != nil {
		return nil, err
	}

	// Configure new default kms key
	_, err = ebs.NewDefaultKmsKey(ctx, name, &ebs.DefaultKmsKeyArgs{
		KeyArn: key.Arn,
	})
	if err != nil {
		return nil, err
	}

	// Ensure that the default kms key is used for all EBS volumes
	_, err = ebs.NewEncryptionByDefault(ctx, name, &ebs.EncryptionByDefaultArgs{
		Enabled: pulumi.Bool(true),
	})
	if err != nil {
		return nil, err
	}

	return key, nil
}

// Our ebs volumes are encrypted by default. This will allow the autoscaling groups to provision a root
// ebs volume for the ec2 instances it spawns.
// https://docs.aws.amazon.com/autoscaling/ec2/userguide/key-policy-requirements-EBS-encryption.html
func DefaultEbsEncryptionKeyPolicy(ctx *pulumi.Context, accountId string) (*iam.GetPolicyDocumentResult, error) {
	effect := "Allow"
	kmsAccess, err := iam.GetPolicyDocument(ctx, &iam.GetPolicyDocumentArgs{
		Statements: []iam.GetPolicyDocumentStatement{
			{
				Effect: &effect,
				Principals: []iam.GetPolicyDocumentStatementPrincipal{
					{
						Type: "AWS",
						Identifiers: []string{
							fmt.Sprintf("arn:aws:iam::%s:root", accountId),
						},

					},
				},
				Actions: []string{
					"kms:*",
				},
				Resources: []string{"*"},
			},
			{
				Effect: &effect,
				Actions: []string{
					"kms:Encrypt",
					"kms:Decrypt",
					"kms:ReEncrypt*",
					"kms:GenerateDataKey*",
					"kms:DescribeKey",
				},
				Principals: []iam.GetPolicyDocumentStatementPrincipal{
					{
						Type: "AWS",
						Identifiers: []string{
							fmt.Sprintf("arn:aws:iam::%s:role/aws-service-role/autoscaling.amazonaws.com/AWSServiceRoleForAutoScaling", accountId),
						},

					},
				},
				Resources: []string{"*"},
			},
			{
				Effect: &effect,
				Actions: []string{
					"kms:CreateGrant",
				},
				Principals: []iam.GetPolicyDocumentStatementPrincipal{
					{
						Type: "AWS",
						Identifiers: []string{
							fmt.Sprintf("arn:aws:iam::%s:role/aws-service-role/autoscaling.amazonaws.com/AWSServiceRoleForAutoScaling", accountId),
						},

					},
				},
				Resources: []string{"*"},
				// allow the user to create grants on the KMS key only when the grant is created on the user's behalf by an AWS service
				// This follows the principle of least priviledge.
				Conditions: []iam.GetPolicyDocumentStatementCondition{
					{
						Test: "Bool",
						Variable: "kms:GrantIsForAWSResource",
						Values: []string{"true"},
					},
				},
			},
		},
	})
	if err != nil { return nil, err }

	return kmsAccess, nil
}
