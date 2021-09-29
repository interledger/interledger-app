package main

import (
	"fmt"

	"github.com/pulumi/pulumi-aws/sdk/v4/go/aws/iam"
	"github.com/pulumi/pulumi-aws/sdk/v4/go/aws/s3"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// This is here if you need to enable logging for the image builder instance. 
func newVaultS3AccessPolicy(ctx *pulumi.Context, bucketArn pulumi.StringOutput) pulumi.StringOutput {
	return pulumi.All(bucketArn).ApplyT(func(args []interface{}) (string, error) {
		bucketArn := args[0].(string)

		s3Access, err := iam.GetPolicyDocument(ctx, &iam.GetPolicyDocumentArgs{
			Statements: []iam.GetPolicyDocumentStatement{
				{
					Effect: stringPtr("Allow"),
					Actions: []string{
						"s3:ListBucket",
					},
					Resources: []string{
						bucketArn,
					},
				},
				{
					Effect: stringPtr("Allow"),
					Actions: []string{
						"s3:PutObject",
						"s3:GetObject",
					},
					Resources: []string{
						fmt.Sprintf("%s/*", bucketArn),
					},
				},
			},
		})
		if err != nil {
			return "", err
		}
		return s3Access.Json, err
	}).(pulumi.StringOutput)
}

func newVaultRoleTrustPolicy(ctx *pulumi.Context) (*iam.GetPolicyDocumentResult, error) {
	policy, err := iam.GetPolicyDocument(ctx, &iam.GetPolicyDocumentArgs{
		Statements: []iam.GetPolicyDocumentStatement{
			{
				Effect: stringPtr("Allow"),
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
	})

	return policy, err
}

func newBuilderThatLogsProfile(ctx *pulumi.Context, bucket *s3.Bucket) (*iam.InstanceProfile, error) {
	trustPolicy, err := newVaultRoleTrustPolicy(ctx)
	if err != nil {
		return nil, err
	}

	role, err := iam.NewRole(ctx, "vault-image-builder-role", &iam.RoleArgs{
		Name: pulumi.String("vault-image-builder"),
		ManagedPolicyArns: pulumi.StringArray{
			pulumi.String("arn:aws:iam::aws:policy/EC2InstanceProfileForImageBuilder"),
			pulumi.String("arn:aws:iam::aws:policy/EC2InstanceProfileForImageBuilderECRContainerBuilds"),
			pulumi.String("arn:aws:iam::aws:policy/AmazonSSMManagedInstanceCore"),
		},
		InlinePolicies: &iam.RoleInlinePolicyArray{
			iam.RoleInlinePolicyArgs{
				Name:   pulumi.String("vaultS3BucketAccessPolicy"),
				Policy: newVaultS3AccessPolicy(ctx, bucket.Arn),
			},
		},
		AssumeRolePolicy: pulumi.String(trustPolicy.Json),
	})
	if err != nil {
		return nil, err
	}

	ec2Profile, err := iam.NewInstanceProfile(ctx, "vault-image-builder-ec2-profile", &iam.InstanceProfileArgs{
		Name: pulumi.String("vault-image-builder"),
		Role: role.Name,
	})
	if err != nil {
		return nil, err
	}

	return ec2Profile, nil
}

func newBuilderProfile(ctx *pulumi.Context) (*iam.InstanceProfile, error) {
	trustPolicy, err := newVaultRoleTrustPolicy(ctx)
	if err != nil {
		return nil, err
	}

	role, err := iam.NewRole(ctx, "vault-image-builder-role", &iam.RoleArgs{
		Name: pulumi.String("vault-image-builder"),
		ManagedPolicyArns: pulumi.StringArray{
			pulumi.String("arn:aws:iam::aws:policy/EC2InstanceProfileForImageBuilder"),
			pulumi.String("arn:aws:iam::aws:policy/EC2InstanceProfileForImageBuilderECRContainerBuilds"),
			pulumi.String("arn:aws:iam::aws:policy/AmazonSSMManagedInstanceCore"),
		},
		AssumeRolePolicy: pulumi.String(trustPolicy.Json),
	})
	if err != nil {
		return nil, err
	}

	ec2Profile, err := iam.NewInstanceProfile(ctx, "vault-image-builder-ec2-profile", &iam.InstanceProfileArgs{
		Name: pulumi.String("vault-image-builder"),
		Role: role.Name,
	})
	if err != nil {
		return nil, err
	}

	return ec2Profile, nil
}

func stringPtr(s string) *string {
	return &s
}
