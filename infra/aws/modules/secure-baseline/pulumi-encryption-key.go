package secure_baseline

import (
	"github.com/pulumi/pulumi-aws/sdk/v4/go/aws/kms"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func NewPulumiSecretsKey (ctx *pulumi.Context, trustedRoleArn pulumi.StringOutput) (*kms.Key, error) {
	key, err := kms.NewKey(ctx, "pulumi-secrets-key", &kms.KeyArgs{
		Description: pulumi.String("Pulumi secrets"),
		Tags:		 pulumi.StringMap{"Name": pulumi.String("Pulumi secrets")},
	})
	if err != nil { return nil, err }

	_, err = kms.NewGrant(ctx, "pulumi-secrets-key-grant", &kms.GrantArgs{
		KeyId:            key.KeyId,
		GranteePrincipal: trustedRoleArn,
		Operations: pulumi.StringArray{
			pulumi.String("Encrypt"),
			pulumi.String("Decrypt"),
			pulumi.String("GenerateDataKey"),
			pulumi.String("DescribeKey"),
		},
	})
	if err != nil { return nil, err }

	return key, nil
}