package secure_baseline

import (
	"github.com/pulumi/pulumi-aws/sdk/v4/go/aws/cloudtrail"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type CloudtrailBaselineArgs struct {
	KmsKeyId     pulumi.StringOutput
	S3BucketName pulumi.StringOutput
}

func NewCloudtrailBaseline(ctx *pulumi.Context, name string, args *CloudtrailBaselineArgs) error {

	// Configure cloudtrail
	_, err := cloudtrail.NewTrail(ctx, name, &cloudtrail.TrailArgs{
		EnableLogFileValidation:    pulumi.Bool(true),
		IncludeGlobalServiceEvents: pulumi.Bool(true),
		IsMultiRegionTrail:         pulumi.Bool(true),
		KmsKeyId:                   args.KmsKeyId,
		S3BucketName:               args.S3BucketName,
	})
	if err != nil {
		return err
	}

	return nil
}
