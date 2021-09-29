package main

import (
	utils "gitlab.com/fynbos/infra/aws/modules/utils"

	"github.com/pulumi/pulumi-aws/sdk/v4/go/aws/ebs"
	"github.com/pulumi/pulumi-aws/sdk/v4/go/aws/kms"
	"github.com/pulumi/pulumi-aws/sdk/v4/go/aws/route53"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		awsConf := config.New(ctx, "aws")
		region := awsConf.Require("region")
		fynbosConf := config.New(ctx, "fynbos")
		accountId := fynbosConf.Require("accountId")
		vpcStack, err := pulumi.NewStackReference(ctx, "fynbos/aws-shared-euwest1-networking/main", nil)
		if err != nil {
			return err
		}
		privateSubnets := utils.StringArrayOutputFromStack(vpcStack, "privateSubnets")
		vpcCidrBlock := vpcStack.GetStringOutput(pulumi.String("vpcCidrBlock"))
		vpcId := vpcStack.GetStringOutput(pulumi.String("vpcId"))
		dnsZoneId := vpcStack.GetIDOutput(pulumi.String("dnsZoneId"))
		
		group, err := newVaultSecurityGroup(ctx, vpcId, vpcCidrBlock)
		if err != nil {
			return err
		}

		key, err := kms.NewKey(ctx, "vault-encryption-key", &kms.KeyArgs{
			Description: pulumi.String("Used to de/encrypt the tls private key."),
		})

		ec2Profile, err := newVaultEc2Profile(ctx, "vault-instance", key.Arn)
		if err != nil {
			return err
		}

		// for the moment we are backing the Vault server with an EBS instance. TODO: move to using raft storage when we deploy more nodes.
		blockDevice, err := ebs.NewVolume(ctx, "vault-ebs-xvdb", &ebs.VolumeArgs{
			AvailabilityZone: pulumi.String("eu-west-1a"),
			Size:             pulumi.Int(50),
			Type: 			      pulumi.String("gp2"),
			Tags: pulumi.StringMap{
				"Name": pulumi.String("vault-xvdb"),
			},
		}, pulumi.Protect(true))
		if err != nil {
			return err
		}
		ctx.Export("vaultEbsId", blockDevice.ID())
		ctx.Export("vaultEbsArn", blockDevice.Arn)

		serialNumber := "0" // Update the serial number when the certicate is rotated /changed. Remember to distribute the Vault cert to clients.
		privateKey, vaultCert, err := newVaultCertificate(ctx, serialNumber)
		if err != nil {
			return err
		}
		ctx.Export("vaultTlsCert", vaultCert.CertPem)

		// encrypt private key as it will be present in the user data which can be queried via instance meta data api.
		encryptedPrivateKey, err := kms.NewCiphertext(ctx, "encrypted-tls-private-key", &kms.CiphertextArgs{
			KeyId: key.ID(),
			Plaintext: privateKey.PrivateKeyPem,
		})

		vault, err := newVaultServer(VaultServerArgs{
			ctx: ctx,
			accountId: accountId,
			ec2Profile: ec2Profile,
			kmsKeyId: key.ID(),
			subnet: privateSubnets.Index(pulumi.Int(0)),
			secGroup: group,
			region: region,
			blockDevice: blockDevice,
			instanceType: "t3.medium",
			encryptedPrivateKey: encryptedPrivateKey.CiphertextBlob,
			tlsCert: vaultCert.CertPem,
		})
		if err != nil {
			return err
		}
		ctx.Export("vaultARN", vault.Arn)

		// create a simple record for vault.fynbos.dev for now as we only have one instance.
		_, err = route53.NewRecord(ctx, "vault.fynbos.cloud", &route53.RecordArgs{
			ZoneId: dnsZoneId,
			Name:   pulumi.String("vault"),
			Type:   pulumi.String("A"),
			Ttl:    pulumi.Int(300),
			Records: pulumi.StringArray{
				vault.PrivateIp,
			},
		})

		return nil
	})
}
