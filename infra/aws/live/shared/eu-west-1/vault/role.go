package main

import (
	"github.com/pulumi/pulumi-aws/sdk/v4/go/aws/iam"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func newVaultIdentityAccessPolicy(ctx *pulumi.Context) string {
	vaultIam, _ := iam.GetPolicyDocument(ctx, &iam.GetPolicyDocumentArgs{
			Statements: []iam.GetPolicyDocumentStatement{
				{
					Effect: stringPtr("Allow"),
					Actions: []string{
						"iam:GetRole",
						"iam:GetUser",
					},
					Resources: []string{
						"arn:aws:iam::*:user/*",
      					"arn:aws:iam::*:role/*",
					},
				},
				{
					Effect: stringPtr("Allow"),
					Actions: []string{
						"sts:GetCallerIdentity",
					},
					Resources: []string{
						"*",
					},
				},
			},
		})

	return vaultIam.Json
}

func newVaultKmsAccessPolicy(ctx *pulumi.Context, keyArn pulumi.StringOutput) pulumi.StringOutput {
	return pulumi.All(keyArn).ApplyT(func(args []interface{}) (string, error) {
		keyArn := args[0].(string)

		kmsAccess, err := iam.GetPolicyDocument(ctx, &iam.GetPolicyDocumentArgs{
			Statements: []iam.GetPolicyDocumentStatement{
				{
					Effect: stringPtr("Allow"),
					Actions: []string{
						"kms:Decrypt",
						"kms:DescribeKey",
						"kms:Encrypt",
						"kms:GenerateDataKey",
					},
					Resources: []string{
						keyArn,
					},
				},
			},
		})

		return kmsAccess.Json, err
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

func stringPtr(s string) *string {
	return &s
}
