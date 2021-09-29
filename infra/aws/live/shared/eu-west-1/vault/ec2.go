package main

import (
	b64 "encoding/base64"
	"gitlab.com/fynbos/infra/aws/modules/utils"

	"github.com/pulumi/pulumi-aws/sdk/v4/go/aws/ebs"
	"github.com/pulumi/pulumi-aws/sdk/v4/go/aws/ec2"
	"github.com/pulumi/pulumi-aws/sdk/v4/go/aws/iam"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func newVaultSecurityGroup (ctx *pulumi.Context, vpcId pulumi.StringOutput,
	vpcCidrBlocks pulumi.StringOutput,
) (*ec2.SecurityGroup, error) {
	group, err := ec2.NewSecurityGroup(ctx, "vault-web-secgrp", &ec2.SecurityGroupArgs{
		VpcId: vpcId,
		Ingress: ec2.SecurityGroupIngressArray{
			// Allow inbound connections to Vault from private and intra subnets
			ec2.SecurityGroupIngressArgs{
				Protocol:   pulumi.String("tcp"),
				FromPort:   pulumi.Int(8200),
				ToPort:     pulumi.Int(8201),
				CidrBlocks: pulumi.StringArray{
					vpcCidrBlocks,
				},
			},
		},
		Egress: ec2.SecurityGroupEgressArray{
			// Allow outbound traffic on 443. Mainly so that it can reach KMS to decrypt private key
			ec2.SecurityGroupEgressArgs{
				FromPort: pulumi.Int(443),
				ToPort:   pulumi.Int(443),
				Protocol: pulumi.String("tcp"),
				CidrBlocks: pulumi.StringArray{
					pulumi.String("0.0.0.0/0"),
				},
			},
			// Allow outbound traffic from Vault to private and intra subnets
			ec2.SecurityGroupEgressArgs{
				FromPort: pulumi.Int(8200),
				ToPort:   pulumi.Int(8201),
				Protocol: pulumi.String("tcp"),
				CidrBlocks: pulumi.StringArray{
					vpcCidrBlocks,
				},
			},
		},
	})
	if err != nil {
		return nil, err
	}

	return group, nil
}

func newVaultEc2Profile(ctx *pulumi.Context,name string, kmsKeyArn pulumi.StringOutput) (*iam.InstanceProfile, error) {
	trustPolicy, err := newVaultRoleTrustPolicy(ctx)
	if err != nil {
		return nil, err
	}

	role, err := iam.NewRole(ctx, name + "-role", &iam.RoleArgs{
		Name: pulumi.String(name),
		InlinePolicies: &iam.RoleInlinePolicyArray{
			iam.RoleInlinePolicyArgs{
				Name:   pulumi.String("vaultKmsAccessPolicy"),
				Policy: newVaultKmsAccessPolicy(ctx, kmsKeyArn),
			},
			iam.RoleInlinePolicyArgs{
				Name:   pulumi.String("vaultIamPolicy"),
				Policy: pulumi.String(newVaultIdentityAccessPolicy(ctx)),
			},
		},
		AssumeRolePolicy: pulumi.String(trustPolicy.Json),
	})
	if err != nil {
		return nil, err
	}

	ec2Profile, err := iam.NewInstanceProfile(ctx, name + "-profile", &iam.InstanceProfileArgs{
		Name: pulumi.String(name),
		Role: role.Name,
	})
	if err != nil {
		return nil, err
	}

	return ec2Profile, nil
}

func vaultCloudInitConfig (encryptedTlsPrivateKey pulumi.StringOutput, tlsCert pulumi.StringOutput, encryptionKeyId pulumi.IDOutput) pulumi.StringOutput {
	type VaultCloudInitData struct {
		EncryptedTlsPrivateKey    string
		TlsCert                   string
		SystemdConfig             string
		VaultConfig               string
		KmsKeyId                  interface{}
	}
	return pulumi.All(encryptedTlsPrivateKey, tlsCert, encryptionKeyId).ApplyT(func(args []interface{}) string {
		data := VaultCloudInitData{
			EncryptedTlsPrivateKey: args[0].(string),
			TlsCert:                b64.StdEncoding.EncodeToString([]byte (args[1].(string))),
			SystemdConfig:          utils.ParseTemplate(VaultCloudInitData{}, "./config/vault.service"),
			VaultConfig:            utils.ParseTemplate(VaultCloudInitData{}, "./config/config.hcl"),
			KmsKeyId:               args[2],
		}

		return utils.ParseTemplate(data, "./cloudinit.yaml")
	}).(pulumi.StringOutput)
}


type VaultServerArgs struct{
	ctx *pulumi.Context
	accountId           string
	kmsKeyId            pulumi.IDOutput
	ec2Profile          *iam.InstanceProfile
	subnet              pulumi.StringOutput
	secGroup            *ec2.SecurityGroup
	region              string
	blockDevice         *ebs.Volume
	instanceType        string
	encryptedPrivateKey pulumi.StringOutput
	tlsCert             pulumi.StringOutput
}
func newVaultServer(opts VaultServerArgs) (*ec2.Instance, error) {
	mostRecent := true
	ami, err := ec2.LookupAmi(opts.ctx, &ec2.LookupAmiArgs{
		Filters: []ec2.GetAmiFilter{
			{
				Name:   "name",
				Values: []string{"fynbos-vault-*"},
			},
		},
		Owners:     []string{opts.accountId},
		MostRecent: &mostRecent,
	})
	if err != nil {
		return nil, err
	}

	vault, err := ec2.NewInstance(opts.ctx, "vault instance", &ec2.InstanceArgs{
		Tags:                pulumi.StringMap{"Name": pulumi.String("vault")},
		InstanceType:        pulumi.String(opts.instanceType),
		VpcSecurityGroupIds: pulumi.StringArray{opts.secGroup.ID()},
		SubnetId:            opts.subnet,
		Ami:                 pulumi.String(ami.Id),
		IamInstanceProfile:  opts.ec2Profile.Name,
		UserDataBase64:      vaultCloudInitConfig(opts.encryptedPrivateKey, opts.tlsCert, opts.kmsKeyId),
	})

	_, err = ec2.NewVolumeAttachment(opts.ctx, "vault-ebs-attachment", &ec2.VolumeAttachmentArgs{
		DeviceName: pulumi.String("/dev/xvdb"),
		VolumeId:   opts.blockDevice.ID(),
		InstanceId: vault.ID(),
	})
	if err != nil {
		return nil, err
	}

	return vault, nil
}