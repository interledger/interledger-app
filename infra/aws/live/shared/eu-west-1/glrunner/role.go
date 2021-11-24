package main

import (
	"encoding/json"

	"github.com/pulumi/pulumi-aws/sdk/v4/go/aws/iam"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// CreateManagerRole
// creates the required role for the gitlab runner manager
// ensures ec2 instances can assume the role
// adds the necessary policies for the role for ec2 management
func CreateManagerRole(ctx *pulumi.Context, ebsKmsKeyArn pulumi.StringOutput, runnerRole *iam.Role) (*iam.Role, error) {
	instanceAssumeRolePolicy, err := iam.GetPolicyDocument(ctx, &iam.GetPolicyDocumentArgs{
		Statements: []iam.GetPolicyDocumentStatement{
			{
				Actions: []string{
					"sts:AssumeRole",
				},
				Principals: []iam.GetPolicyDocumentStatementPrincipal{
					{
						Type: "Service",
						Identifiers: []string{
							"ec2.amazonaws.com",
						},
					},
				},
			},
		},
	}, nil)
	if err != nil {
		return nil, err
	}

	role, err := iam.NewRole(ctx, "gl-runner-manager", &iam.RoleArgs{
		Path:             pulumi.String("/"),
		AssumeRolePolicy: pulumi.String(instanceAssumeRolePolicy.Json),
		ManagedPolicyArns: pulumi.StringArray{
			pulumi.String("arn:aws:iam::aws:policy/AmazonEC2ContainerRegistryPowerUser"),
		},
		InlinePolicies: iam.RoleInlinePolicyArray{
			iam.RoleInlinePolicyArgs{
				Name:   pulumi.String("ebsKmsKeyAccessPolicy"),
				Policy: ebsEncryptionKeyAccessPolicy(ctx, ebsKmsKeyArn),
			},
			iam.RoleInlinePolicyArgs{
				Name:   pulumi.String("manageRunnerEc2InstancesPolicy"),
				Policy: manageEc2Policy(),
			},
			iam.RoleInlinePolicyArgs{
				Name:   pulumi.String("passRunnerRolePolicy"),
				Policy: passRolePolicy(ctx, runnerRole),
			},
		},
	})
	if err != nil {
		return nil, err
	}

	return role, nil
}

// Creates role which allows EC2 instances to push/pull from private ECR repos.
func CreateRunnerRole(ctx *pulumi.Context) (*iam.Role, error) {
	instanceAssumeRolePolicy, err := iam.GetPolicyDocument(ctx, &iam.GetPolicyDocumentArgs{
		Statements: []iam.GetPolicyDocumentStatement{
			{
				Actions: []string{
					"sts:AssumeRole",
				},
				Principals: []iam.GetPolicyDocumentStatementPrincipal{
					{
						Type: "Service",
						Identifiers: []string{
							"ec2.amazonaws.com",
						},
					},
				},
			},
		},
	}, nil)
	if err != nil {
		return nil, err
	}

	role, err := iam.NewRole(ctx, "gl-runner-instance", &iam.RoleArgs{
		Path:             pulumi.String("/"),
		AssumeRolePolicy: pulumi.String(instanceAssumeRolePolicy.Json),
		ManagedPolicyArns: pulumi.StringArray{
			pulumi.String("arn:aws:iam::aws:policy/AmazonEC2ContainerRegistryPowerUser"),
		},
	})
	if err != nil {
		return nil, err
	}

	return role, nil
}

func passRolePolicy(ctx *pulumi.Context, role *iam.Role) pulumi.StringOutput {
	effect := "Allow"
	return pulumi.All(role.Arn).ApplyT(func(args []interface{}) (string, error) {
		roleArn := args[0].(string)

		passRole, err := iam.GetPolicyDocument(ctx, &iam.GetPolicyDocumentArgs{
			Statements: []iam.GetPolicyDocumentStatement{
				{
					Effect: &effect,
					Actions: []string{
						"iam:PassRole",
					},
					Resources: []string{
						roleArn,
					},
				},
			},
		})
		if err != nil {
			return "", err
		}

		return passRole.Json, err
	}).(pulumi.StringOutput)
}

func manageEc2Policy() pulumi.String {

	type Statement struct {
		Effect   string   `json:"Effect"`
		Action   []string `json:"Action,omitempty"`
		Resource string   `json:"Resource"`
	}

	type Policy struct {
		Version    string      `json:"Version"`
		ID         string      `json:"Id"`
		Statements []Statement `json:"Statement"`
	}

	rawPolicy := &Policy{
		Version: "2012-10-17",
		ID:      "Gitlab runner management instance policy",
		Statements: []Statement{
			{
				Effect: "Allow",
				Action: []string{
					"ec2:DescribeSubnets",
					"ec2:DescribeSecurityGroups",
					"ec2:DescribeKeyPairs",
					"ec2:ImportKeyPair",
					"ec2:DeleteKeyPair",
					"ec2:RequestSpotInstances",
					"ec2:DescribeSpotInstanceRequests",
					"ec2:CancelSpotInstanceRequests",
					"ec2:RunInstances",
					"ec2:DescribeInstances",
					"ec2:CreateTags",
					"ec2:TerminateInstances",
				},
				Resource: "*",
			},
		},
	}

	policy, _ := json.Marshal(rawPolicy)

	return pulumi.String(policy)
}

// Our ebs volumes are encrypted by default. This will allow the provisioning of a root
// ebs volume for ec2 instances.
// https://docs.aws.amazon.com/autoscaling/ec2/userguide/key-policy-requirements-EBS-encryption.html
func ebsEncryptionKeyAccessPolicy(ctx *pulumi.Context, keyArn pulumi.StringOutput) pulumi.StringOutput {
	effect := "Allow"
	return pulumi.All(keyArn).ApplyT(func(args []interface{}) (string, error) {
		keyArn := args[0].(string)

		kmsAccess, err := iam.GetPolicyDocument(ctx, &iam.GetPolicyDocumentArgs{
			Statements: []iam.GetPolicyDocumentStatement{
				{
					Effect: &effect,
					Actions: []string{
						"kms:Encrypt",
						"kms:Decrypt",
						"kms:ReEncrypt*",
						"kms:GenerateDataKey*",
						"kms:DescribeKey",
					},
					Resources: []string{
						keyArn,
					},
				},
				{
					Effect: &effect,
					Actions: []string{
						"kms:CreateGrant",
					},
					Resources: []string{
						keyArn,
					},
					// allow the user to create grants on the KMS key only when the grant is created on the user's behalf by an AWS service
					// This follows the principle of least priviledge.
					Conditions: []iam.GetPolicyDocumentStatementCondition{
						{
							Test:     "Bool",
							Variable: "kms:GrantIsForAWSResource",
							Values:   []string{"true"},
						},
					},
				},
			},
		})
		if err != nil {
			return "", err
		}

		return kmsAccess.Json, err
	}).(pulumi.StringOutput)
}
