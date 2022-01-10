package kubernetes

import (
	"github.com/pulumi/pulumi-aws/sdk/v4/go/aws/iam"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"gitlab.com/fynbos/infra/aws/modules/utils"
)

type EbsCsiArgs struct {
	EbsKmsKeyArn pulumi.StringOutput
}

func DeployEbsCsi(ctx *pulumi.Context, args EbsCsiArgs) error {

	// Create Policy/ServiceAccount
	_, err := newPolicy(ctx, args.ebsKmsKeyArn)

	//_, err := eks.NewAddon(ctx, "aws-ebs-csi", &eks.AddonArgs{
	//	ClusterName: pulumi.Any(aws_eks_cluster.Example.Name),
	//	AddonName:   pulumi.String("aws-ebs-csi-driver"),
	//})
	if err != nil {
		return err
	}

	return nil
}

/**
Generates a policy for CSI to be able to manage volumes for EKS as well as Encrypt and Decrypt with KMS key for EBS
See https://docs.aws.amazon.com/eks/latest/userguide/managing-ebs-csi.html for more info
*/
func newPolicy(ctx *pulumi.Context, ebsKmsKey pulumi.StringArrayOutput) (*iam.Policy, error) {

	policyDoc, err := iam.GetPolicyDocument(ctx, &iam.GetPolicyDocumentArgs{
		Version: utils.StringPtr("2012-10-17"),
		Statements: []iam.GetPolicyDocumentStatement{
			{
				Effect: utils.StringPtr("Allow"),
				Actions: []string{
					"ec2:CreateSnapshot",
					"ec2:AttachVolume",
					"ec2:DetachVolume",
					"ec2:ModifyVolume",
					"ec2:DescribeAvailabilityZones",
					"ec2:DescribeInstances",
					"ec2:DescribeSnapshots",
					"ec2:DescribeTags",
					"ec2:DescribeVolumes",
					"ec2:DescribeVolumesModifications",
				},
				Resources: []string{"*"},
			},
			{
				Effect: utils.StringPtr("Allow"),
				Actions: []string{
					"ec2:CreateTags",
				},
				Resources: []string{
					"arn:aws:ec2:*:*:volume/*",
					"arn:aws:ec2:*:*:snapshot/*",
				},
				Conditions: []iam.GetPolicyDocumentStatementCondition{
					{
						Test:     "StringEquals",
						Variable: "ec2:CreateAction",
						Values: []string{
							"CreateVolume",
							"CreateSnapshot",
						},
					},
				},
			},
			{
				Effect: utils.StringPtr("Allow"),
				Actions: []string{
					"ec2:DeleteTags",
				},
				Resources: []string{
					"arn:aws:ec2:*:*:volume/*",
					"arn:aws:ec2:*:*:snapshot/*",
				},
			},
			{
				Effect: utils.StringPtr("Allow"),
				Actions: []string{
					"ec2:CreateVolume",
				},
				Resources: []string{
					"*",
				},
				Conditions: []iam.GetPolicyDocumentStatementCondition{
					{
						Test:     "StringLike",
						Variable: "aws:RequestTag/ebs.csi.aws.com/cluster",
						Values: []string{
							"true",
						},
					},
				},
			},
			{
				Effect: utils.StringPtr("Allow"),
				Actions: []string{
					"ec2:CreateVolume",
				},
				Resources: []string{
					"*",
				},
				Conditions: []iam.GetPolicyDocumentStatementCondition{
					{
						Test:     "StringLike",
						Variable: "aws:RequestTag/CSIVolumeName",
						Values: []string{
							"*",
						},
					},
				},
			},
			{
				Effect: utils.StringPtr("Allow"),
				Actions: []string{
					"ec2:CreateVolume",
				},
				Resources: []string{
					"*",
				},
				Conditions: []iam.GetPolicyDocumentStatementCondition{
					{
						Test:     "StringLike",
						Variable: "aws:RequestTag/kubernetes.io/cluster/*",
						Values: []string{
							"owned",
						},
					},
				},
			},
			{
				Effect: utils.StringPtr("Allow"),
				Actions: []string{
					"ec2:DeleteVolume",
				},
				Resources: []string{
					"*",
				},
				Conditions: []iam.GetPolicyDocumentStatementCondition{
					{
						Test:     "StringLike",
						Variable: "ec2:ResourceTag/ebs.csi.aws.com/cluster",
						Values: []string{
							"true",
						},
					},
				},
			},
			{
				Effect: utils.StringPtr("Allow"),
				Actions: []string{
					"ec2:DeleteVolume",
				},
				Resources: []string{
					"*",
				},
				Conditions: []iam.GetPolicyDocumentStatementCondition{
					{
						Test:     "StringLike",
						Variable: "aws:RequestTag/CSIVolumeName",
						Values: []string{
							"*",
						},
					},
				},
			},
			{
				Effect: utils.StringPtr("Allow"),
				Actions: []string{
					"ec2:DeleteVolume",
				},
				Resources: []string{
					"*",
				},
				Conditions: []iam.GetPolicyDocumentStatementCondition{
					{
						Test:     "StringLike",
						Variable: "aws:RequestTag/kubernetes.io/cluster/*",
						Values: []string{
							"owned",
						},
					},
				},
			},
			{
				Effect: utils.StringPtr("Allow"),
				Actions: []string{
					"ec2:DeleteSnapshot",
				},
				Resources: []string{
					"*",
				},
				Conditions: []iam.GetPolicyDocumentStatementCondition{
					{
						Test:     "StringLike",
						Variable: "ec2:ResourceTag/CSIVolumeSnapshotName",
						Values: []string{
							"*",
						},
					},
				},
			},
			{
				Effect: utils.StringPtr("Allow"),
				Actions: []string{
					"ec2:DeleteSnapshot",
				},
				Resources: []string{
					"*",
				},
				Conditions: []iam.GetPolicyDocumentStatementCondition{
					{
						Test:     "StringLike",
						Variable: "ec2:ResourceTag/ebs.csi.aws.com/cluster",
						Values: []string{
							"true",
						},
					},
				},
			},
			{
				Effect: utils.StringPtr("Allow"),
				Actions: []string{
					"kms:CreateGrant",
					"kms:ListGrants",
					"kms:RevokeGrant",
				},
				Resources: []string{
					ebsKmsKey,
				},
				Conditions: []iam.GetPolicyDocumentStatementCondition{
					{
						Test:     "Bool",
						Variable: "kms:GrantIsForAWSResource",
						Values: []string{
							"true",
						},
					},
				},
			},
			{
				Effect: utils.StringPtr("Allow"),
				Actions: []string{
					"kms:Encrypt",
					"kms:Decrypt",
					"kms:ReEncrypt*",
					"kms:GenerateDataKey*",
					"kms:DescribeKey",
				},
				Resources: []string{
					ebsKmsKey,
				},
			},
		},
	})
	if err != nil {
		return nil, err
	}

	policy, err := iam.NewPolicy(ctx, "aws-csi-ebs", &iam.PolicyArgs{
		Name:        pulumi.String("AmazonEKSEBSCSIDriverPolicy"),
		Description: pulumi.String("This policy provides the Amazon EBS Add on permission to manage EBS volumes and encryption"),
		Path:        pulumi.String("/"),
		Policy:      pulumi.String(policyDoc.Json),
	})
	if err != nil {
		return nil, err
	}

	return policy, nil
}
