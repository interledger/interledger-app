package main

import (
	b64 "encoding/base64"
	"encoding/json"
	"fmt"

	"github.com/pulumi/pulumi-aws/sdk/v4/go/aws/ec2"
	"github.com/pulumi/pulumi-aws/sdk/v4/go/aws/eks"
	"github.com/pulumi/pulumi-aws/sdk/v4/go/aws/iam"
	"github.com/pulumi/pulumi-aws/sdk/v4/go/aws/kms"
	"github.com/pulumi/pulumi-aws/sdk/v4/go/aws/s3"
	"github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes"
	corev1 "github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes/core/v1"
	metav1 "github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes/meta/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"
	k8s "gitlab.com/fynbos/infra/aws/modules/kubernetes"
	"gitlab.com/fynbos/infra/aws/modules/utils"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		accountID := "634848879735"
		clusterName := "dev-eu-west-1-cluster"
		awsConfig := config.New(ctx, "aws")
		region := awsConfig.Get("region")
		vpcStack, err := pulumi.NewStackReference(ctx, "fynbos/aws-dev-euwest1-networking/main", nil)
		if err != nil {
			return err
		}
		publicSubnetIds := utils.StringArrayOutputFromStack(vpcStack, "publicSubnets")
		privateSubnetIds := utils.StringArrayOutputFromStack(vpcStack, "privateSubnets")
		key, err := kms.NewKey(ctx, "cluster-encryption-key", &kms.KeyArgs{
			Tags: pulumi.StringMap{
				"name": pulumi.Sprintf("%s-encryption-key", clusterName),
			},
		}, pulumi.Protect(false))
		if err != nil {
			return err
		}

		//Concat the subnets
		subnets := pulumi.All(publicSubnetIds, privateSubnetIds).ApplyT(func(args []interface{}) ([]string, error) {
			publicSubnets := args[0].([]string)
			privateSubnets := args[1].([]string)
			subnets := append(publicSubnets, privateSubnets...)
			return subnets, nil
		}).(pulumi.StringArrayOutput)

		clusterRole, err := k8s.NewServiceRole(ctx, fmt.Sprintf("%s-cluster-role", clusterName), &k8s.ServiceRoleArgs{
			Service:     "eks.amazonaws.com",
			Description: "Allows EKS to manage clusters on your behalf.",
			ManagedPolicyArns: []string{
				"arn:aws:iam::aws:policy/AmazonEKSClusterPolicy",
			},
		})
		if err != nil {
			return err
		}

		cluster, err := eks.NewCluster(ctx, "dev-cluster", &eks.ClusterArgs{
			RoleArn: clusterRole.Arn,
			Name:    pulumi.String(clusterName),
			Version: pulumi.String("1.26"),
			VpcConfig: eks.ClusterVpcConfigArgs{
				EndpointPrivateAccess: pulumi.BoolPtr(true),
				EndpointPublicAccess:  pulumi.BoolPtr(true),
				SubnetIds:             subnets,
			},
			EncryptionConfig: eks.ClusterEncryptionConfigArgs{
				Provider: eks.ClusterEncryptionConfigProviderArgs{
					KeyArn: key.Arn,
				},
				Resources: pulumi.StringArray{
					pulumi.String("secrets"),
				},
			},
			EnabledClusterLogTypes: pulumi.StringArray{
				pulumi.String("api"),
				pulumi.String("audit"),
				pulumi.String("authenticator"),
				pulumi.String("controllerManager"),
				pulumi.String("scheduler"),
			},
			Tags: pulumi.StringMap{
				"Name": pulumi.String(clusterName),
			},
		}, pulumi.DependsOn([]pulumi.Resource{key, clusterRole}))
		if err != nil {
			return err
		}
		ctx.Export("clusterEndpoint", cluster.Endpoint)
		ctx.Export("ca", cluster.CertificateAuthority.Data())

		// Create Kubeconfig
		kubeConfig := pulumi.All(cluster.Name, cluster.CertificateAuthority, cluster.Endpoint).ApplyT(func(args []interface{}) (string, error) {
			name := args[0].(string)
			ca := args[1].(eks.ClusterCertificateAuthority).Data
			endpoint := args[2].(string)
			c := k8s.GenerateKubeconfig(name, ca, endpoint)
			j, err := json.Marshal(c)
			return string(j), err
		}).(pulumi.StringOutput)
		ctx.Export("kubeconfig", kubeConfig)

		//Create OIDC Provider
		thumbprint := k8s.FingerprintAddress(fmt.Sprintf("https://oidc.eks.%s.amazonaws.com/", region))
		provider, err := iam.NewOpenIdConnectProvider(ctx, "oidc-provider", &iam.OpenIdConnectProviderArgs{
			Url: cluster.Identities.Index(pulumi.Int(0)).Oidcs().Index(pulumi.Int(0)).Issuer().Elem(),
			ThumbprintLists: pulumi.StringArray{
				pulumi.String(thumbprint),
			},
			ClientIdLists: pulumi.StringArray{
				pulumi.String("sts.amazonaws.com"),
			},
		}, pulumi.DependsOn([]pulumi.Resource{cluster}))

		ctx.Export("oidcProvider", provider.Url)

		instanceRole, err := k8s.NewServiceRole(ctx, fmt.Sprintf("%s-instance-role", clusterName), &k8s.ServiceRoleArgs{
			Service:     "ec2.amazonaws.com",
			Description: "Allows nodes to operate in cluster.",
			ManagedPolicyArns: []string{
				"arn:aws:iam::aws:policy/AmazonEKSWorkerNodePolicy",
				"arn:aws:iam::aws:policy/AmazonEKS_CNI_Policy", // TODO check if this is required
				"arn:aws:iam::aws:policy/AmazonEC2ContainerRegistryReadOnly",
			},
		})
		if err != nil {
			return err
		}

		deployRole, err := k8s.NewDeployRole(ctx, clusterName, "arn:aws:iam::823058932981:role/eksArgoRole")
		if err != nil {
			return err
		}
		ctx.Export("deployRoleArn", deployRole.Arn)

		kubeProvider, err := kubernetes.NewProvider(ctx, "kubernetes-provider", &kubernetes.ProviderArgs{
			Kubeconfig: kubeConfig,
		})
		if err != nil {
			return err
		}

		//roleConfig := k8s.RoleMappingConfig([]*iam.Role{instanceRole}, []k8s.RoleMap{})
		roleConfig := k8s.RoleMappingConfigV2([]k8s.RoleMapArg{
			{
				RoleArn:  instanceRole.Arn,
				Username: pulumi.String("system:node:{{EC2PrivateDNSName}}"),
				Groups: pulumi.StringArray{
					pulumi.String("system:bootstrappers"),
					pulumi.String("system:nodes"),
				},
			},
			{
				RoleArn:  deployRole.Arn,
				Username: deployRole.Arn,
				Groups: pulumi.StringArray{
					pulumi.String("system:masters"),
				},
			},
		})

		authConfig, err := corev1.NewConfigMap(ctx, "aws-auth", &corev1.ConfigMapArgs{
			Metadata: &metav1.ObjectMetaArgs{
				Name:      pulumi.String("aws-auth"),
				Namespace: pulumi.String("kube-system"),
			},
			Data: pulumi.StringMap{
				"mapRoles": roleConfig,
			},
		}, pulumi.Provider(kubeProvider), pulumi.DependsOn([]pulumi.Resource{
			instanceRole,
			cluster,
			deployRole,
		}))
		if err != nil {
			return err
		}

		err = createGeneralNodeGroup(ctx, CreateNodeGroupArgs{
			Cluster:      cluster,
			InstanceRole: instanceRole,
			SubnetIds:    privateSubnetIds,
			AuthConfig:   authConfig,
		})
		if err != nil {
			return err
		}

		// crdb backups bucket
		crdbBackupsBucket, err := s3.NewBucket(ctx, "crdb-backups-bucket", &s3.BucketArgs{
			Versioning: &s3.BucketVersioningArgs{
				Enabled: pulumi.Bool(true),
			},
			LifecycleRules: s3.BucketLifecycleRuleArray{
				s3.BucketLifecycleRuleArgs{
					Expiration: s3.BucketLifecycleRuleExpirationArgs{
						Days: pulumi.Int(30),
					},
					Enabled: pulumi.Bool(true),
				},
			},
			ServerSideEncryptionConfiguration: s3.BucketServerSideEncryptionConfigurationArgs{
				Rule: s3.BucketServerSideEncryptionConfigurationRuleArgs{
					BucketKeyEnabled: pulumi.Bool(true), // use bucket level kms key
					ApplyServerSideEncryptionByDefault: s3.BucketServerSideEncryptionConfigurationRuleApplyServerSideEncryptionByDefaultArgs{
						SseAlgorithm: pulumi.String("aws:kms"),
					},
				},
			},
		}, pulumi.Provider(kubeProvider), pulumi.Protect(true))
		if err != nil {
			return err
		}
		ctx.Export("crdbBackupsBucket", crdbBackupsBucket.Arn)

		backupTrustPolicy := k8s.NewIamTrustPolicyDocumentV2(ctx, pulumi.String(accountID), provider.Url, pulumi.String("cockroachdb"), pulumi.String("cockroachdb"))
		backupAccessPolicy := pulumi.All(crdbBackupsBucket.Arn).ApplyT(func(args []interface{}) (string, error) {
			bucketARN := args[0].(string)

			policy, err := k8s.NewBucketReadWriteDeleteAccessPolicy(ctx, bucketARN)
			if err != nil {
				return "", err
			}

			return policy.Json, nil
		}).(pulumi.StringOutput)

		backupRole, err := iam.NewRole(ctx, "crdb-backup", &iam.RoleArgs{
			AssumeRolePolicy: backupTrustPolicy,
			InlinePolicies: iam.RoleInlinePolicyArray{
				iam.RoleInlinePolicyArgs{
					Name:   pulumi.String("read-write-access"),
					Policy: backupAccessPolicy,
				},
			},
		}, pulumi.Provider(kubeProvider))
		if err != nil {
			return err
		}
		ctx.Export("crdbBackupsRole", backupRole.Arn)

		readBackupTrustPolicy := k8s.NewIamTrustPolicyDocumentV2(ctx, pulumi.String(accountID), provider.Url, pulumi.String("cockroachbackupcheck"), pulumi.String("cockroachdb"))
		readBackupAccessPolicy := pulumi.All(crdbBackupsBucket.Arn).ApplyT(func(args []interface{}) (string, error) {
			bucketARN := args[0].(string)

			policy, err := k8s.NewBucketReadOnlyAccessPolicy(ctx, bucketARN)
			if err != nil {
				return "", err
			}

			return policy.Json, nil
		}).(pulumi.StringOutput)

		readBackupRole, err := iam.NewRole(ctx, "crdb-read-backup", &iam.RoleArgs{
			AssumeRolePolicy: readBackupTrustPolicy,
			InlinePolicies: iam.RoleInlinePolicyArray{
				iam.RoleInlinePolicyArgs{
					Name:   pulumi.String("read-access"),
					Policy: readBackupAccessPolicy,
				},
			},
		}, pulumi.Provider(kubeProvider))
		if err != nil {
			return err
		}
		ctx.Export("crdbReadBackupsRole", readBackupRole.Arn)

		return nil
	})
}

type CreateNodeGroupArgs struct {
	Cluster      *eks.Cluster
	InstanceRole *iam.Role
	SubnetIds    pulumi.StringArrayOutput
	AuthConfig   *corev1.ConfigMap
}

func createGeneralNodeGroup(ctx *pulumi.Context, args CreateNodeGroupArgs) error {
	instanceUserData := `
[settings.kubernetes]
max-pods = 40
`
	// Launch template is required to set values in the user data field. This
	// is merged with the default values from the EKS Node group to create
	// a new launch template
	launchTemplate, err := ec2.NewLaunchTemplate(ctx, "eks-general-template-1", &ec2.LaunchTemplateArgs{
		UserData: pulumi.String(b64.StdEncoding.EncodeToString([]byte(instanceUserData))),
		BlockDeviceMappings: ec2.LaunchTemplateBlockDeviceMappingArray{
			ec2.LaunchTemplateBlockDeviceMappingArgs{
				DeviceName: pulumi.String("/dev/xvda"),
				Ebs: &ec2.LaunchTemplateBlockDeviceMappingEbsArgs{
					VolumeType: pulumi.String("gp3"),
					VolumeSize: pulumi.Int(2),
					Throughput: pulumi.Int(125),
					Iops:       pulumi.Int(3000),
				},
			},
			ec2.LaunchTemplateBlockDeviceMappingArgs{
				DeviceName: pulumi.String("/dev/xvdb"),
				Ebs: &ec2.LaunchTemplateBlockDeviceMappingEbsArgs{
					VolumeType: pulumi.String("gp3"),
					VolumeSize: pulumi.Int(20),
					Throughput: pulumi.Int(125),
					Iops:       pulumi.Int(3000),
				},
			},
		},
	})
	if err != nil {
		return err
	}

	_, err = eks.NewNodeGroup(ctx, "managed-general-1", &eks.NodeGroupArgs{
		NodeGroupName: pulumi.String("managed-general-1"),
		ClusterName:   args.Cluster.Name,
		ScalingConfig: eks.NodeGroupScalingConfigArgs{
			DesiredSize: pulumi.Int(3),
			MinSize:     pulumi.Int(3),
			MaxSize:     pulumi.Int(4),
		},
		NodeRoleArn: args.InstanceRole.Arn,
		LaunchTemplate: eks.NodeGroupLaunchTemplateArgs{
			Id:      launchTemplate.ID(),
			Version: pulumi.Sprintf("%d", launchTemplate.LatestVersion),
		},
		InstanceTypes: pulumi.StringArray{pulumi.String("t3.medium")},
		SubnetIds:     args.SubnetIds,
		AmiType:       pulumi.String("BOTTLEROCKET_x86_64"),
		Taints: eks.NodeGroupTaintArray{
			eks.NodeGroupTaintArgs{
				Effect: pulumi.String("NO_EXECUTE"),
				Key:    pulumi.String("node.cilium.io/agent-not-ready"),
				Value:  pulumi.String("true"),
			},
		},
	}, pulumi.DependsOn([]pulumi.Resource{launchTemplate, args.Cluster, args.InstanceRole, args.AuthConfig}))
	if err != nil {
		return err
	}

	return nil
}
