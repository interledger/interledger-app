package secure_baseline

import (
	"github.com/pulumi/pulumi-aws/sdk/v4/go/aws/s3"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func NewS3AccountBaseline(ctx *pulumi.Context) error {
	_, err := s3.NewAccountPublicAccessBlock(ctx, "awsS3AccountPublicAccessBlock", &s3.AccountPublicAccessBlockArgs{
		BlockPublicAcls:       pulumi.Bool(true),
		BlockPublicPolicy:     pulumi.Bool(true),
		IgnorePublicAcls:      pulumi.Bool(true),
		RestrictPublicBuckets: pulumi.Bool(true),
	})
	if err != nil {
		return err
	}
	return nil
}
