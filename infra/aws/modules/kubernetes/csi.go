package kubernetes

import (
	"github.com/pulumi/pulumi-aws/sdk/v4/go/aws/eks"
	"github.com/pulumi/pulumi-aws/sdk/v4/go/aws/iam"
	metav1 "github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes/meta/v1"
	storagev1 "github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes/storage/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"gitlab.com/fynbos/infra/aws/modules/utils"
)

type EbsCsiArgs struct {
	EbsKmsKeyArn pulumi.StringOutput
	OidcProvider string
	AccountId    string
	ClusterName  string
}

func DeployEbsCsi(ctx *pulumi.Context, args EbsCsiArgs) error {

	// Create Policy
	policy, err := newPolicy(ctx, args.EbsKmsKeyArn, args.ClusterName)
	if err != nil {
		return err
	}

	// Create Trust Policy
	trustPolicy, err := NewIamTrustPolicyDocument(ctx, args.AccountId, args.OidcProvider, "kube-system", "ebs-csi-controller-sa")
	if err != nil {
		return err
	}

	// Role
	role, err := iam.NewRole(ctx, "eks-ebs-csi-role", &iam.RoleArgs{
		Name: pulumi.Sprintf("%s-EBSCSIRole", args.ClusterName),
		ManagedPolicyArns: pulumi.StringArray{
			policy.Arn,
		},
		AssumeRolePolicy: pulumi.String(trustPolicy),
		Description:      pulumi.String("IAM Role that has policy to manage EBS and KMS for EKS"),
		Path:             pulumi.String("/"),
	})
	if err != nil {
		return err
	}

	_, err = eks.NewAddon(ctx, "aws-ebs-csi-addon", &eks.AddonArgs{
		ClusterName:           pulumi.String(args.ClusterName),
		AddonName:             pulumi.String("aws-ebs-csi-driver"),
		ServiceAccountRoleArn: role.Arn,
	})
	if err != nil {
		return err
	}

	return nil
}

/**
Generates a policy for CSI to be able to manage volumes for EKS as well as Encrypt and Decrypt with KMS key for EBS
See https://docs.aws.amazon.com/eks/latest/userguide/managing-ebs-csi.html for more info
*/
func newPolicy(ctx *pulumi.Context, ebsKmsKey pulumi.StringOutput, clusterName string) (*iam.Policy, error) {

	policyDoc := pulumi.All(ebsKmsKey).ApplyT(func(args []interface{}) (string, error) {
		key := args[0].(string)

		doc, err := iam.GetPolicyDocument(ctx, &iam.GetPolicyDocumentArgs{
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
						key,
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
						key,
					},
				},
			},
		})
		return doc.Json, err
	})

	policy, err := iam.NewPolicy(ctx, "aws-csi-ebs", &iam.PolicyArgs{
		Name:        pulumi.Sprintf("%sEbsCsiDriverPolicy", clusterName),
		Description: pulumi.String("This policy provides the Amazon EBS Add on permission to manage EBS volumes and encryption"),
		Path:        pulumi.String("/"),
		Policy:      policyDoc,
	})
	if err != nil {
		return nil, err
	}

	return policy, nil
}

func DeployDefaultCSIStorageClass(ctx *pulumi.Context, opts ...pulumi.ResourceOption) error {
	_, err := storagev1.NewStorageClass(ctx, "ebs-sc-default", &storagev1.StorageClassArgs{
		Provisioner:       pulumi.String("ebs.csi.aws.com"),
		VolumeBindingMode: pulumi.String("WaitForFirstConsumer"),
		Metadata: metav1.ObjectMetaArgs{
			Name: pulumi.String("ebs-sc"),
			Annotations: pulumi.StringMap{
				"storageclass.kubernetes.io/is-default-class": pulumi.String("true"),
			},
		},
	}, opts...)
	if err != nil {
		return err
	}

	return nil
}
