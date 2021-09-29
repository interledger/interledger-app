package main

import (
	"github.com/pulumi/pulumi-aws/sdk/v4/go/aws/ec2"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"
	"gitlab.com/fynbos/infra/aws/modules/utils"
)

func main() {

	pulumi.Run(func(ctx *pulumi.Context) error {
		awsConf := config.New(ctx, "aws")
		region := awsConf.Require("region")
		vpcStack, err := pulumi.NewStackReference(ctx, "fynbos/aws-shared-euwest1-networking/main", nil)
		if err != nil {
			return err
		}
		privateSubnets := utils.StringArrayOutputFromStack(vpcStack, "privateSubnets")
		sg, err := ec2.NewSecurityGroup(ctx, "vault-web-secgrp", &ec2.SecurityGroupArgs{
			Egress: ec2.SecurityGroupEgressArray{
				ec2.SecurityGroupEgressArgs{ // allow outbound so that it can download Vault binary
					FromPort: pulumi.Int(0),
					ToPort:   pulumi.Int(0),
					Protocol: pulumi.String("-1"),
					CidrBlocks: pulumi.StringArray{
						pulumi.String("0.0.0.0/0"),
					},
					Ipv6CidrBlocks: pulumi.StringArray{
						pulumi.String("::/0"),
					},
				},
			},
			VpcId: vpcStack.GetStringOutput(pulumi.String("vpcId")),
			Tags: pulumi.StringMap{
				"service": pulumi.String("vault-image-builder"),
			},
		})
		if err != nil {
			return err
		}

		infrastructure, distribution, logsBucket, err := newImageBuilderConfiguration(ImageBuilderConfig{
			ctx: ctx,
			subnet: privateSubnets.Index(pulumi.Int(0)),
			region: region,
			secGroup: sg,
			enableLogging: false,
		})
		if err != nil {
			return err
		}
		if logsBucket != nil {			
			ctx.Export("vaultBucketId", logsBucket.ID())
			ctx.Export("vaultBucketArn", logsBucket.Arn)
		}

		image, err := newVaultAmi(ctx, infrastructure.Arn, distribution.Arn, "0.0.25")
		if err != nil {
			return err
		}
		ctx.Export("vaultImageArn", image.Arn)

		return nil
	})
}
