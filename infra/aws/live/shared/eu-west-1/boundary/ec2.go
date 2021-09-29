package main

import (
	"net/url"

	"github.com/pulumi/pulumi-aws/sdk/v4/go/aws/ec2"
	"github.com/pulumi/pulumi-aws/sdk/v4/go/aws/iam"
	"github.com/pulumi/pulumi-aws/sdk/v4/go/aws/kms"
	"github.com/pulumi/pulumi-aws/sdk/v4/go/aws/rds"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	utils "gitlab.com/fynbos/infra/aws/modules/utils"

	b64 "encoding/base64"
)

func createKeys(ctx *pulumi.Context) (*kms.Key, *kms.Key, *kms.Key, error) {
	recoveryKey, err := kms.NewKey(ctx, "boundary-recovery-key", &kms.KeyArgs{
		Description: pulumi.String("Boundary recovery key"),
	})
	if err != nil {
		return nil, nil, nil, err
	}
	_, err = kms.NewAlias(ctx, "boundary-recovery-key-alias", &kms.AliasArgs{
		TargetKeyId: recoveryKey.KeyId,
		Name: pulumi.String("alias/boundary-recovery"),
	})
	if err != nil {
		return nil, nil, nil, err
	}

	rootKey, err := kms.NewKey(ctx, "boundary-root-key", &kms.KeyArgs{
		Description: pulumi.String("Boundary root key"),
	})
	if err != nil {
		return nil, nil, nil, err
	}

	workerKey, err := kms.NewKey(ctx, "boundary-worker-key", &kms.KeyArgs{
		Description: pulumi.String("Boundary worker key"),
	})
	if err != nil {
		return nil, nil, nil, err
	}

	return recoveryKey, rootKey, workerKey, nil
}

func createControllerSecurityGroup(ctx *pulumi.Context, vpcId pulumi.StringOutput, privateSubnetsCidrBlocks pulumi.StringArrayOutput) (*ec2.SecurityGroup, error) {
	return ec2.NewSecurityGroup(ctx, "boundary-controller", &ec2.SecurityGroupArgs{
		Name:        pulumi.String("boundary-controller"),
		VpcId:       vpcId,
		Description: pulumi.String("Boundary Controller"),
		Ingress: ec2.SecurityGroupIngressArray{
			&ec2.SecurityGroupIngressArgs{ // controller api endpoint
				FromPort: pulumi.Int(9200),
				ToPort:   pulumi.Int(9200),
				Protocol: pulumi.String("tcp"),
				CidrBlocks: pulumi.StringArray{
					pulumi.String("0.0.0.0/0"), // allow from anywhere as users will use cli on their local machines. i.e. network load balancer does not adjust src ip address.
				},
			},
			&ec2.SecurityGroupIngressArgs{ // allow inbound communication from worker
				FromPort: pulumi.Int(9201),
				ToPort:   pulumi.Int(9201),
				Protocol: pulumi.String("tcp"),
				CidrBlocks: privateSubnetsCidrBlocks,
			},
		},
		Egress: ec2.SecurityGroupEgressArray{
			// Allow outbound communication to postgres.
			&ec2.SecurityGroupEgressArgs{
				FromPort: pulumi.Int(5432),
				ToPort:   pulumi.Int(5432),
				Protocol: pulumi.String("tcp"),
				CidrBlocks: pulumi.StringArray{
					pulumi.String("0.0.0.0/0"),
				},
			},
			&ec2.SecurityGroupEgressArgs{ // allow https out so that instance can speak to AWS KMS and install Boundary
				FromPort: pulumi.Int(443),
				ToPort:   pulumi.Int(443),
				Protocol: pulumi.String("tcp"),
				CidrBlocks: pulumi.StringArray{
					pulumi.String("0.0.0.0/0"),
				},
			},
			&ec2.SecurityGroupEgressArgs{ // Allow outbound to Vault
				FromPort: pulumi.Int(8200),
				ToPort:   pulumi.Int(8200),
				Protocol: pulumi.String("tcp"),
				CidrBlocks: privateSubnetsCidrBlocks,
			},
			&ec2.SecurityGroupEgressArgs{ // Allow outbound to Worker
				FromPort: pulumi.Int(9202),
				ToPort:   pulumi.Int(9202),
				Protocol: pulumi.String("tcp"),
				CidrBlocks: privateSubnetsCidrBlocks,
			},
		},
	})
}

func createWorkerSecurityGroup(ctx *pulumi.Context, vpcId pulumi.StringOutput, vpcCidrBlock pulumi.StringOutput) (*ec2.SecurityGroup, error) {
	return ec2.NewSecurityGroup(ctx, "boundary-worker-sg", &ec2.SecurityGroupArgs{
		Name:        pulumi.String("boundary-worker"),
		VpcId:       vpcId,
		Description: pulumi.String("Boundary Worker Sg"),
		Ingress: ec2.SecurityGroupIngressArray{
			// allow inbound whilst acting as proxy
			&ec2.SecurityGroupIngressArgs{
				FromPort: pulumi.Int(9202),
				ToPort:   pulumi.Int(9202),
				Protocol: pulumi.String("tcp"),
				CidrBlocks: pulumi.StringArray{
					pulumi.String("0.0.0.0/0"), // allow from anywhere as users will use cli on their local machines. i.e. network load balancer does not adjust src ip address.
				},
			},
		},
		Egress: ec2.SecurityGroupEgressArray{
			// Allow outbound communication to entire VPC
			&ec2.SecurityGroupEgressArgs{
				FromPort: pulumi.Int(0),
				ToPort:   pulumi.Int(65535),
				Protocol: pulumi.String("tcp"),
				CidrBlocks: pulumi.StringArray{
					vpcCidrBlock,
				},
			},
			&ec2.SecurityGroupEgressArgs{ // allow https out so that instance can speak to AWS KMS and install Boundary
				FromPort: pulumi.Int(443),
				ToPort:   pulumi.Int(443),
				Protocol: pulumi.String("tcp"),
				CidrBlocks: pulumi.StringArray{
					pulumi.String("0.0.0.0/0"),
				},
			},
		},
	})
}

type CloudInitData struct {
	BoundaryType              string
	Config                    string
	InstallScript             string
	TlsCert                   string
	EncryptedTlsPrivateKey    string
	KmsKeyId                  interface{} // The worker key will be used to en/decrypt the tls private key
}
func workerCloudInitData(workerKeyId pulumi.IDOutput, controllerIps pulumi.StringArray, publicHost pulumi.StringOutput, tlsCert pulumi.StringOutput, encryptedTlsPrivateKey pulumi.StringOutput) pulumi.StringOutput {
	type WorkerConfigData struct {
		TlsDisabled    bool
		KmsWorkerKeyId interface{}
		ControllerIps  []string
		PublicHost     string
	}

	return pulumi.All(workerKeyId, controllerIps, publicHost, tlsCert, encryptedTlsPrivateKey).ApplyT(func(args []interface{}) (string, error) {
		config := WorkerConfigData{
			KmsWorkerKeyId: args[0],
			ControllerIps:  args[1].([]string),
			PublicHost:     args[2].(string),
			TlsDisabled:    false,
		}

		data := CloudInitData{
			BoundaryType: "worker",
			Config:        utils.ParseTemplate(config, "./install/worker.hcl"),
			InstallScript: utils.ParseTemplate(config, "./install/install.sh"),
			TlsCert: 	   b64.StdEncoding.EncodeToString([]byte(args[3].(string))),
			EncryptedTlsPrivateKey:  args[4].(string),
			KmsKeyId:      args[0],
		}

		return utils.ParseTemplate(data, "./install/cloudinit.yaml"), nil
	}).(pulumi.StringOutput)
}

func controllerCloudInitData(dbEndpoint pulumi.StringOutput, dbPassword pulumi.StringOutput, recoveryKeyId pulumi.IDOutput, rootKeyArn pulumi.IDOutput, workerKeyId pulumi.IDOutput, tlsCert pulumi.StringOutput, encryptedTlsPrivateKey pulumi.StringOutput) pulumi.StringOutput {
	type ControllerConfigData struct {
		DbEndpoint       string
		DbPassword       string
		KmsRecoveryKeyId interface{}
		KmsRootKeyId     interface{}
		KmsWorkerKeyId   interface{}
		TlsDisabled      bool
	}
	return pulumi.All(dbEndpoint, dbPassword, recoveryKeyId, rootKeyArn, workerKeyId, tlsCert, encryptedTlsPrivateKey).ApplyT(func(args []interface{}) (string, error) {
		config := ControllerConfigData{
			DbEndpoint: args[0].(string),
			DbPassword: url.QueryEscape(args[1].(string)), // url encode so that net/url (boundary uses it to parse connection string) will parse it correctly
			KmsRecoveryKeyId: args[2],
			KmsRootKeyId: args[3],
			KmsWorkerKeyId: args[4],
			TlsDisabled: false,
		}

		data := CloudInitData{
			BoundaryType: "controller",
			Config:        utils.ParseTemplate(config, "./install/controller.hcl"),
			InstallScript: utils.ParseTemplate(config, "./install/install.sh"),
			TlsCert: b64.StdEncoding.EncodeToString([]byte(args[5].(string))),
			EncryptedTlsPrivateKey: args[6].(string),
			KmsKeyId: args[4],
		}

		return utils.ParseTemplate(data, "./install/cloudinit.yaml"), nil
	}).(pulumi.StringOutput)
}

func createWorkerEc2Profile(ctx *pulumi.Context, workerKeyArn pulumi.StringOutput) (*iam.InstanceProfile, error) {
	trustPolicy, err := newBoundaryRoleTrustPolicy(ctx)
	if err != nil {
		return nil, err
	}

	role, err := iam.NewRole(ctx, "boundary-worker-role", &iam.RoleArgs{
		AssumeRolePolicy: pulumi.String(trustPolicy.Json),
		Description: pulumi.String("boundary-worker-role"),
		InlinePolicies: iam.RoleInlinePolicyArray{
			iam.RoleInlinePolicyArgs{
				Name:   pulumi.String("kms-access-policy"),
				Policy: newBoundaryWorkerKmsAccessPolicy(ctx, workerKeyArn),
			},

		},
	})
	if err != nil {
		return nil, err
	}

	ec2Profile, err := iam.NewInstanceProfile(ctx, "boundary-worker-profile", &iam.InstanceProfileArgs{
		Name: pulumi.String("boundary-worker-profile"),
		Role: role,
	})
	if err != nil {
		return nil, err
	}

	return ec2Profile, nil
}

func createControllerEc2Profile(ctx *pulumi.Context, recoveryKeyArn pulumi.StringOutput, rootKeyArn pulumi.StringOutput, workerKeyArn pulumi.StringOutput) (*iam.InstanceProfile, error) {
	trustPolicy, err := newBoundaryRoleTrustPolicy(ctx)
	if err != nil {
		return nil, err
	}

	role, err := iam.NewRole(ctx, "boundary-controller-role", &iam.RoleArgs{
		AssumeRolePolicy: pulumi.String(trustPolicy.Json),
		Description: pulumi.String("boundary-controller-role"),
		InlinePolicies: iam.RoleInlinePolicyArray{
			iam.RoleInlinePolicyArgs{
				Name:   pulumi.String("kms-access-policy"),
				Policy: newBoundaryControllerKmsAccessPolicy(ctx, recoveryKeyArn, rootKeyArn, workerKeyArn),
			},

		},
	})
	if err != nil {
		return nil, err
	}

	ec2Profile, err := iam.NewInstanceProfile(ctx, "boundary-controller-profile", &iam.InstanceProfileArgs{
		Name: pulumi.String("boundary-controller-profile"),
		Role: role,
	})
	if err != nil {
		return nil, err
	}

	return ec2Profile, nil
}

type CreateWorkerArgs struct {
	ctx                    *pulumi.Context
	subnetId               pulumi.StringOutput
	sg                     *ec2.SecurityGroup
	workerKey              *kms.Key
	controllerIps          pulumi.StringArray
	ec2Profile             *iam.InstanceProfile
	publicHost             pulumi.StringOutput
	tlsCert                pulumi.StringOutput
	encryptedTlsPrivateKey pulumi.StringOutput
}
func createWorkerInstance(args CreateWorkerArgs) (*ec2.Instance, error) {
	opt0 := true
	ami, err := ec2.LookupAmi(args.ctx, &ec2.LookupAmiArgs{
		MostRecent: &opt0,
		Owners: []string{
			"amazon",
		},
		Filters: []ec2.GetAmiFilter{
			{
				Name: "name",
				Values: []string{
					"amzn2-ami-hvm-*-x86_64-gp2",
				},
			},
		},
	})
	if err != nil {
		return nil, err
	}

	instance, err := ec2.NewInstance(args.ctx, "boundary-worker", &ec2.InstanceArgs{
		Ami:          pulumi.String(ami.Id),
		InstanceType: pulumi.String("t3.nano"),
		SubnetId:     args.subnetId,
		VpcSecurityGroupIds: pulumi.StringArray{
			args.sg.ID(),
		},
		UserDataBase64: workerCloudInitData(args.workerKey.ID(), args.controllerIps, args.publicHost, args.tlsCert, args.encryptedTlsPrivateKey),
		IamInstanceProfile: args.ec2Profile,
		Tags:         pulumi.StringMap{"Name": pulumi.String("Boundary Worker")},
	})

	return instance, nil
}

type CreateControllerArgs struct {
	ctx                      *pulumi.Context
	subnetId                 pulumi.StringOutput
	sg                       *ec2.SecurityGroup
	recoveryKey              *kms.Key
	rootKey                  *kms.Key
	workerKey                *kms.Key
	postgres    	         *rds.Instance
	dbPassword               pulumi.StringOutput
	ec2Profile               *iam.InstanceProfile
	tlsCert                  pulumi.StringOutput
	encryptedTlsPrivateKey   pulumi.StringOutput
}
func createControllerInstance(args CreateControllerArgs) (*ec2.Instance, error) {
	opt0 := true
	ami, err := ec2.LookupAmi(args.ctx, &ec2.LookupAmiArgs{
		MostRecent: &opt0,
		Owners: []string{
			"amazon",
		},
		Filters: []ec2.GetAmiFilter{
			{
				Name: "name",
				Values: []string{
					"amzn2-ami-hvm-*-x86_64-gp2",
				},
			},
		},
	})
	if err != nil {
		return nil, err
	}

	instance, err := ec2.NewInstance(args.ctx, "boundary-controller", &ec2.InstanceArgs{
		Ami:          pulumi.String(ami.Id),
		InstanceType: pulumi.String("t3.nano"),
		SubnetId:     args.subnetId,
		VpcSecurityGroupIds: pulumi.StringArray{
			args.sg.ID(),
		},
		UserDataBase64: controllerCloudInitData(args.postgres.Endpoint, args.dbPassword, args.recoveryKey.ID(), args.rootKey.ID(), args.workerKey.ID(), args.tlsCert, args.encryptedTlsPrivateKey),
		IamInstanceProfile: args.ec2Profile,
		Tags:         pulumi.StringMap{"Name": pulumi.String("Boundary Controller")},
	})

	return instance, nil
}