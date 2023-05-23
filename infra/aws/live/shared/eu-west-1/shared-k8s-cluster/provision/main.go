package main

import (
	b64 "encoding/base64"
	"encoding/json"
	"fmt"

	"github.com/pulumi/pulumi-aws/sdk/v4/go/aws/ec2"
	"github.com/pulumi/pulumi-aws/sdk/v4/go/aws/eks"
	"github.com/pulumi/pulumi-aws/sdk/v4/go/aws/iam"
	"github.com/pulumi/pulumi-aws/sdk/v4/go/aws/kms"
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
		clusterName := "shared-eu-west-1-cluster"
		awsConfig := config.New(ctx, "aws")
		region := awsConfig.Get("region")
		// Setting this to `true` will publically expose the clusters admin endpoint.
		// We do this so that we can configure the cluster afterwhich we will disable
		// the public access.
		//allowPublicConfiguration := os.Getenv("ALLOW_PUBLIC_CONFIGURATION") == "true"
		//fynbosConf := config.New(ctx, "fynbos")
		//awsAccountId := fynbosConf.Get("accountId")
		vpcStack, err := pulumi.NewStackReference(ctx, "fynbos/aws-shared-euwest1-networking/main", nil)
		if err != nil {
			return err
		}
		//vpcId := vpcStack.GetIDOutput(pulumi.String("vpcId"))
		publicSubnetIds := utils.StringArrayOutputFromStack(vpcStack, "publicSubnets")
		privateSubnetIds := utils.StringArrayOutputFromStack(vpcStack, "privateSubnets")
		//glRunnerStack, err := pulumi.NewStackReference(ctx, "fynbos/aws-shared-euwest1-glrunner/main", nil)
		//if err != nil {
		//	return err
		//}

		//glRunnerRoleArn := glRunnerStack.GetStringOutput(pulumi.String("glRunnerRoleArn"))
		//glRunnerSg := glRunnerStack.GetIDOutput(pulumi.String("glRunnerSecurityGroupID"))
		//
		//roles, err := k8s.NewEksRoles(ctx,"shared", awsAccountId, glRunnerRoleArn)
		//if err != nil {
		//	return err
		//}
		//ctx.Export("adminRoleArn", roles.Admin.Arn)
		//ctx.Export("automationRoleArn", roles.Automation.Arn)

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

		cluster, err := eks.NewCluster(ctx, "cluster", &eks.ClusterArgs{
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

		kubeProvider, err := kubernetes.NewProvider(ctx, "kubernetes-provider", &kubernetes.ProviderArgs{
			Kubeconfig: kubeConfig,
		})
		if err != nil {
			return err
		}

		deployRole, err := k8s.NewDeployRole(ctx, clusterName, "arn:aws:iam::823058932981:role/eksArgoRole")
		if err != nil {
			return err
		}
		ctx.Export("deployRoleArn", deployRole.Arn)

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

		err = createVaultNodeGroup(ctx, CreateNodeGroupArgs{
			Cluster:      cluster,
			InstanceRole: instanceRole,
			SubnetIds:    privateSubnetIds,
			AuthConfig:   authConfig,
		})
		if err != nil {
			return err
		}

		err = createGitlabRunnerNodeGroup(ctx, CreateNodeGroupArgs{
			Cluster:      cluster,
			InstanceRole: instanceRole,
			SubnetIds:    privateSubnetIds,
			AuthConfig:   authConfig,
		})
		if err != nil {
			return err
		}

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
	launchTemplate, err := ec2.NewLaunchTemplate(ctx, "eks-settings-template-1", &ec2.LaunchTemplateArgs{
		UserData: pulumi.String(b64.StdEncoding.EncodeToString([]byte(instanceUserData))),
	})
	if err != nil {
		return err
	}

	_, err = eks.NewNodeGroup(ctx, "managed-ng-1", &eks.NodeGroupArgs{
		NodeGroupName: pulumi.String("managed-ng-1"),
		ClusterName:   args.Cluster.Name,
		ScalingConfig: eks.NodeGroupScalingConfigArgs{
			DesiredSize: pulumi.Int(2),
			MinSize:     pulumi.Int(1),
			MaxSize:     pulumi.Int(3),
		},
		NodeRoleArn: args.InstanceRole.Arn,
		LaunchTemplate: eks.NodeGroupLaunchTemplateArgs{
			Id:      launchTemplate.ID(),
			Version: pulumi.Sprintf("%d", launchTemplate.LatestVersion),
		},
		InstanceTypes: pulumi.StringArray{pulumi.String("t2.medium")},
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

func createVaultNodeGroup(ctx *pulumi.Context, args CreateNodeGroupArgs) error {
	instanceUserData := `
[settings.kubernetes]
max-pods = 20
`
	// Launch template is required to set values in the user data field. This
	// is merged with the default values from the EKS Node group to create
	// a new launch template
	launchTemplate, err := ec2.NewLaunchTemplate(ctx, "eks-vault-template", &ec2.LaunchTemplateArgs{
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

	_, err = eks.NewNodeGroup(ctx, "managed-vault", &eks.NodeGroupArgs{
		NodeGroupName: pulumi.String("managed-vault"),
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
			eks.NodeGroupTaintArgs{
				Effect: pulumi.String("NO_EXECUTE"),
				Key:    pulumi.String("taint_for_consul_xor_vault"),
				Value:  pulumi.String("true"),
			},
		},
		Labels: pulumi.StringMap{
			"vault_in_k8s": pulumi.String("true"),
		},
	}, pulumi.DependsOn([]pulumi.Resource{launchTemplate, args.Cluster, args.InstanceRole, args.AuthConfig}))
	if err != nil {
		return err
	}

	return nil
}

func createGitlabRunnerNodeGroup(ctx *pulumi.Context, args CreateNodeGroupArgs) error {
	launchTemplate, err := ec2.NewLaunchTemplate(ctx, "eks-glrunner-template", &ec2.LaunchTemplateArgs{
		BlockDeviceMappings: ec2.LaunchTemplateBlockDeviceMappingArray{
			ec2.LaunchTemplateBlockDeviceMappingArgs{
				DeviceName: pulumi.String("/dev/xvdb"),
				Ebs: &ec2.LaunchTemplateBlockDeviceMappingEbsArgs{
					VolumeType: pulumi.String("gp3"),
					VolumeSize: pulumi.Int(100),
					Throughput: pulumi.Int(125),
					Iops:       pulumi.Int(3000),
				},
			},
		},
	})
	if err != nil {
		return err
	}

	_, err = eks.NewNodeGroup(ctx, "managed-glrunner", &eks.NodeGroupArgs{
		NodeGroupName: pulumi.String("managed-glrunner"),
		ClusterName:   args.Cluster.Name,
		ScalingConfig: eks.NodeGroupScalingConfigArgs{
			DesiredSize: pulumi.Int(1),
			MinSize:     pulumi.Int(1),
			MaxSize:     pulumi.Int(2),
		},
		NodeRoleArn: args.InstanceRole.Arn,
		LaunchTemplate: eks.NodeGroupLaunchTemplateArgs{
			Id:      launchTemplate.ID(),
			Version: pulumi.Sprintf("%d", launchTemplate.LatestVersion),
		},
		InstanceTypes: pulumi.StringArray{
			pulumi.String("c6a.2xlarge"),
			pulumi.String("m6a.2xlarge"),
		},
		CapacityType: pulumi.String("SPOT"),
		SubnetIds:    args.SubnetIds,
		AmiType:      pulumi.String("BOTTLEROCKET_x86_64"),
		Taints: eks.NodeGroupTaintArray{
			eks.NodeGroupTaintArgs{
				Effect: pulumi.String("NO_EXECUTE"),
				Key:    pulumi.String("node.cilium.io/agent-not-ready"),
				Value:  pulumi.String("true"),
			},
			eks.NodeGroupTaintArgs{
				Effect: pulumi.String("NO_EXECUTE"),
				Key:    pulumi.String("taint_for_gl_runner"),
				Value:  pulumi.String("true"),
			},
		},
		Labels: pulumi.StringMap{
			"glrunner_in_k8s": pulumi.String("true"),
		},
	}, pulumi.DependsOn([]pulumi.Resource{launchTemplate, args.Cluster, args.InstanceRole, args.AuthConfig}))
	if err != nil {
		return err
	}

	return nil
}
