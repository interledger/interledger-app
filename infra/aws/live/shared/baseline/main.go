package main

import (
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"
	secure_baseline "gitlab.com/fynbos/infra/aws/modules/secure-baseline"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		conf := config.New(ctx, "fynbos")
		securityAccountId := conf.Require("securityAccountId")
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
		err = secure_baseline.NewCloudtrailBaseline(ctx, "shared-ct", &secure_baseline.CloudtrailBaselineArgs{
			KmsKeyId:     pulumi.Sprintf("%s", ctKMSKeyArn),
			S3BucketName: pulumi.Sprintf("%s", ctS3BucketName),
		})
		if err != nil {
			return err
		}

		// invoke default encryption on ebs volumes
		_, err = secure_baseline.EbsBaseline(ctx, "shared-ebs")
		if err != nil {
			return err
		}

		role, err := secure_baseline.NewCrossAccountIamRoles(ctx, securityAccountId)
		if err != nil {
			return err
		}

		pulumiKey, err := secure_baseline.NewPulumiSecretsKey(ctx, role.Arn)
		if err != nil {
			return err
		}

		ctx.Export("pulumiSecretsKeyId", pulumiKey.ID())
		ctx.Export("pulumiSecretsKeyArn", pulumiKey.Arn)

		return nil
	})
}
