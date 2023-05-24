package main

import (
	"fmt"
	"github.com/pulumi/pulumi-aws/sdk/v4/go/aws/iam"
	"github.com/pulumi/pulumi-aws/sdk/v4/go/aws/kms"
	"github.com/pulumi/pulumi-aws/sdk/v4/go/aws/s3"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		key, err := kms.NewKey(ctx, "tabapay-s3-key", &kms.KeyArgs{
			Description:          pulumi.String("This key is used to encrypt tabapay objects"),
			DeletionWindowInDays: pulumi.Int(10),
		})
		if err != nil {
			return err
		}

		bucket, err := s3.NewBucket(ctx, "tabapay-s3", &s3.BucketArgs{
			Bucket: pulumi.String("fynbos-tabapay"),
			Acl:    pulumi.String("private"),
			ServerSideEncryptionConfiguration: &s3.BucketServerSideEncryptionConfigurationArgs{
				Rule: &s3.BucketServerSideEncryptionConfigurationRuleArgs{
					ApplyServerSideEncryptionByDefault: &s3.BucketServerSideEncryptionConfigurationRuleApplyServerSideEncryptionByDefaultArgs{
						KmsMasterKeyId: key.Arn,
						SseAlgorithm:   pulumi.String("aws:kms"),
					},
				},
			},
		})
		if err != nil {
			return err
		}

		p, err := policy(ctx, bucket)
		if err != nil {
			return err
		}

		_, err = s3.NewBucketPolicy(ctx, "tabapay-bucket-policy", &s3.BucketPolicyArgs{
			Bucket: bucket.ID(),
			Policy: p,
		})
		if err != nil {
			return err
		}

		ctx.Export("bucket", bucket.Arn)

		return nil
	})
}

func policy(ctx *pulumi.Context, bucket *s3.Bucket) (pulumi.StringPtrOutput, error) {
	document := iam.GetPolicyDocumentOutput(ctx, iam.GetPolicyDocumentOutputArgs{
		Statements: iam.GetPolicyDocumentStatementArray{
			&iam.GetPolicyDocumentStatementArgs{
				Principals: iam.GetPolicyDocumentStatementPrincipalArray{
					&iam.GetPolicyDocumentStatementPrincipalArgs{
						Type: pulumi.String("AWS"),
						Identifiers: pulumi.StringArray{
							pulumi.String("arn:aws:iam::952934770203:user/tabapay"),
						},
					},
				},
				Actions: pulumi.StringArray{
					pulumi.String("s3:GetObject"),
					pulumi.String("s3:ListBucket"),
					pulumi.String("s3:PutObject"),
				},
				Resources: pulumi.StringArray{
					bucket.Arn,
					bucket.Arn.ApplyT(func(arn string) (string, error) {
						return fmt.Sprintf("%v/*", arn), nil
					}).(pulumi.StringOutput),
				},
			},
		},
	}, nil)

	return document.ApplyT(func(documentPolicy iam.GetPolicyDocumentResult) (*string, error) {
		return &documentPolicy.Json, nil
	}).(pulumi.StringPtrOutput), nil
}
