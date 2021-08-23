package main

import (
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	secure_baseline "gitlab.com/fynbos/infra/aws/modules/secure-baseline"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		rootStack, err := pulumi.NewStackReference(ctx, "fynbos/aws-root/baseline", nil)
		if err != nil {
			return err
		}
		ctKMSKeyArn := rootStack.GetOutput(pulumi.String("cloudtrailKMSKeyArn"))
		ctS3BucketName := rootStack.GetOutput(pulumi.String("cloudtrailS3BucketName"))

		// Configure blocking s3 public
		err = secure_baseline.NewS3AccountBaseline(ctx)
		if err != nil {
			return err
		}

		// Configure cloudtrail
		err = secure_baseline.NewCloudtrailBaseline(ctx, "logs-ct", &secure_baseline.CloudtrailBaselineArgs{
			KmsKeyId:     pulumi.Sprintf("%s", ctKMSKeyArn),
			S3BucketName: pulumi.Sprintf("%s", ctS3BucketName),
		})
		if err != nil {
			return err
		}

		return nil
	})
}
