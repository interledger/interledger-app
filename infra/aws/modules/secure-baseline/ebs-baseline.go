package secure_baseline

import (
	"github.com/pulumi/pulumi-aws/sdk/v4/go/aws/ebs"
	"github.com/pulumi/pulumi-aws/sdk/v4/go/aws/kms"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func EbsBaseline(ctx *pulumi.Context, name string) (*kms.Key, error) {

	key, err := kms.NewKey(ctx, "ebs", &kms.KeyArgs{
		Description: pulumi.String("KMS key ebs baseline"),
	})
	if err != nil {
		return nil, err
	}

	// Configure new default kms key
	_, err = ebs.NewDefaultKmsKey(ctx, name, &ebs.DefaultKmsKeyArgs{
		KeyArn: key.Arn,
	})
	if err != nil {
		return nil, err
	}

	// Ensure that the default kms key is used for all EBS volumes
	_, err = ebs.NewEncryptionByDefault(ctx, name, &ebs.EncryptionByDefaultArgs{
		Enabled: pulumi.Bool(true),
	})
	if err != nil {
		return nil, err
	}

	return key, nil
}

