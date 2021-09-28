package main

import (
	"github.com/pulumi/pulumi-aws/sdk/v4/go/aws/iam"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func newBoundaryControllerKmsAccessPolicy(ctx *pulumi.Context, recoveryKeyArn pulumi.StringOutput, rootKeyArn pulumi.StringOutput, workerkeyArn pulumi.StringOutput) pulumi.StringOutput {
	return pulumi.All(recoveryKeyArn, rootKeyArn, workerkeyArn).ApplyT(func(args []interface{}) (string, error) {
		kmsAccess, err := iam.GetPolicyDocument(ctx, &iam.GetPolicyDocumentArgs{
			Statements: []iam.GetPolicyDocumentStatement{
				{
					Effect: stringPtr("Allow"),
					Actions: []string{
						"kms:DescribeKey",
					    "kms:GenerateDataKey",
					    "kms:Decrypt",
					    "kms:Encrypt",
					},
					Resources: []string{
						args[0].(string),
						args[1].(string),
						args[2].(string),
					},
				},{
					Effect: stringPtr("Allow"),
					Actions: []string{						
					    "kms:ListKeys",
					    "kms:ListAliases",
					},
					Resources: []string{
						"*",
					},
				},
			},
		})

		return kmsAccess.Json, err
	}).(pulumi.StringOutput)
}

func newBoundaryWorkerKmsAccessPolicy(ctx *pulumi.Context, workerkeyArn pulumi.StringOutput) pulumi.StringOutput {
	return pulumi.All(workerkeyArn).ApplyT(func(args []interface{}) (string, error) {
		kmsAccess, err := iam.GetPolicyDocument(ctx, &iam.GetPolicyDocumentArgs{
			Statements: []iam.GetPolicyDocumentStatement{
				{
					Effect: stringPtr("Allow"),
					Actions: []string{
						"kms:DescribeKey",
					    "kms:GenerateDataKey",
					    "kms:Decrypt",
					    "kms:Encrypt",
					},
					Resources: []string{
						args[0].(string),
					},
				},{
					Effect: stringPtr("Allow"),
					Actions: []string{						
					    "kms:ListKeys",
					    "kms:ListAliases",
					},
					Resources: []string{
						"*",
					},
				},
			},
		})

		return kmsAccess.Json, err
	}).(pulumi.StringOutput)
}

func newBoundaryRoleTrustPolicy(ctx *pulumi.Context) (*iam.GetPolicyDocumentResult, error) {
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