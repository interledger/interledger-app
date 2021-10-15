package kubernetes

import (
	"github.com/pulumi/pulumi-aws/sdk/v4/go/aws/iam"
	"github.com/pulumi/pulumi-aws/sdk/v4/go/aws/kms"
	"github.com/pulumi/pulumi-eks/sdk/go/eks"
	policyV1 "github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes/policy/v1beta1"
	rbacV1 "github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes/rbac/v1"
	metaV1 "github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes/meta/v1"
	coreV1 "github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes/core/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type EksControlPlaneArgs struct {
	Version 		string
	Name 			string
	PublicSubnets 	pulumi.StringArrayOutput
	PrivateSubnets 	pulumi.StringArrayOutput
	VpcId			pulumi.IDOutput
	AccountId       string
	IamRoles        EksIamRoles
	ExposeAdminEndpoint bool
}
func NewEksControlPlane(ctx *pulumi.Context, args EksControlPlaneArgs) (*eks.Cluster, error) {
	// This key will be used to create an encrypted envelope around K8s secrets and config.
	// Although k8s already generates its own encryption key to encrypt data it stores in etcd, we use envelope encryption
	// as it is recommended security best practice.
	// **NB** It's impossible to recover the cluster if this key is deleted. We therefore add Pulumi protection.
	key, err := kms.NewKey(ctx, "cluster-encryption-key", &kms.KeyArgs{
		Tags: pulumi.StringMap{
			"name": pulumi.String(args.Name + "-encryption-key"),
		},
	}, pulumi.Protect(true))
	if err != nil { return nil, err }

	cluster, err := eks.NewCluster(ctx, args.Name, &eks.ClusterArgs{
		Name: 						  pulumi.String(args.Name),
		Version: 					  pulumi.String(args.Version),
		VpcId:                        args.VpcId,
		PublicSubnetIds:              args.PublicSubnets, // Pulumi will manage the tagging of the subnets
		PrivateSubnetIds:             args.PrivateSubnets,
		// **NB** The role used to create the cluster (allow-full-access-from-other-accounts in this case) is mapped to the k8s 
		// cluster-admin automatically and won't show up in the role mappings. We will not override this as we want to be able to
		// manage eks provisioning using pulumi.
		RoleMappings: eks.RoleMappingArray{
			// Provides full administrator cluster access to the k8s cluster
			eks.RoleMappingArgs{
				RoleArn: args.IamRoles.Admin.Arn,
				Groups: pulumi.StringArray{pulumi.String("system:masters")},
				Username: args.IamRoles.Admin.Arn,
			},
			// Map IAM automation role arn to the k8s automation group. The role bindings are set up in `ConfigureAutomationRole`.
			eks.RoleMappingArgs{
				RoleArn: args.IamRoles.Automation.Arn,
				Groups: pulumi.StringArray{pulumi.String("automation")},
				Username: args.IamRoles.Automation.Arn,
			},
		},
		StorageClasses: eks.StorageClassArgs{ // Defaults to immediate volume binding
			Type: 					pulumi.String("gp2"),
			Encrypted: 				pulumi.Bool(true),
			Default: 				pulumi.Bool(true),
		},
		SkipDefaultNodeGroup: 		pulumi.Bool(true),	// We will manage the node group separately from the control plane
		EndpointPublicAccess: 		pulumi.Bool(args.ExposeAdminEndpoint),
		EndpointPrivateAccess: 		pulumi.Bool(true),
		NodeAssociatePublicIpAddress: pulumi.Bool(false),
		InstanceRoles: 				iam.RoleArray{      // IAM roles to register with the cluster auth.
			args.IamRoles.NodeGroup,
		},
		NodeSecurityGroupTags:     pulumi.StringMap{ 	// Run with default node security group for now.
			"Name": pulumi.String("eksNodeSecurityGroup"),
		},
		ClusterSecurityGroupTags: pulumi.StringMap{ 	// Run with default cluster security group for now. This will enable full internet egress and ingress from node groups
			"Name": pulumi.String("eksNodeSecurityGroup"),
		},
		EnabledClusterLogTypes: pulumi.StringArray{
			pulumi.String("api"),
			pulumi.String("audit"),
			pulumi.String("authenticator"),
			pulumi.String("controllerManager"),
			pulumi.String("scheduler"),
		},
		EncryptionConfigKeyArn: key.Arn, // enable envelope encyption https://aws.amazon.com/about-aws/whats-new/2020/03/amazon-eks-adds-envelope-encryption-for-secrets-with-aws-kms/
		CreateOidcProvider: pulumi.Bool(true),
	})
	if err != nil { return nil, err }	

	return cluster, nil
}

// This sets up a rescticted pod security policy and applies it to service accounts, using cluster role bindings,
// running in the kube-system namespace as well as the automation group.
func ConfigureClusterRolesAndPsp(ctx *pulumi.Context) error {
	_, err := policyV1.NewPodSecurityPolicy(ctx, "resitricted-psp", &policyV1.PodSecurityPolicyArgs{
		Metadata: metaV1.ObjectMetaArgs{
			Name: pulumi.String("restricted"),
		},
		Spec: policyV1.PodSecurityPolicySpecArgs{
			Privileged: pulumi.Bool(false),
			AllowPrivilegeEscalation: pulumi.Bool(false),
			DefaultAllowPrivilegeEscalation: pulumi.Bool(false),
			HostPID: pulumi.Bool(false),
			HostIPC: pulumi.Bool(false),
			HostNetwork: pulumi.Bool(false),
			Volumes: pulumi.StringArray{
				pulumi.String("configMap"),
				pulumi.String("emptyDir"),
				pulumi.String("projected"),
				pulumi.String("secret"),
				pulumi.String("downwardAPI"),
				pulumi.String("persistentVolumeClaim"),
			},
			RequiredDropCapabilities: pulumi.StringArray{
				pulumi.String("ALL"),
			},
			RunAsUser: policyV1.RunAsUserStrategyOptionsArgs{
				Rule: pulumi.String("MustRunAsNonRoot"),
			},
			SeLinux: policyV1.SELinuxStrategyOptionsArgs{
				Rule: pulumi.String("RunAsAny"),
			},
			SupplementalGroups: policyV1.SupplementalGroupsStrategyOptionsArgs{
				Rule: pulumi.String("MustRunAs"),
				Ranges: policyV1.IDRangeArray{
					policyV1.IDRangeArgs{
						Min: pulumi.Int(1),
						Max: pulumi.Int(65535),
					},
				},
			},
			FsGroup: policyV1.FSGroupStrategyOptionsArgs{
				Rule: pulumi.String("MustRunAs"),
				Ranges: policyV1.IDRangeArray{
					policyV1.IDRangeArgs{
						Min: pulumi.Int(1),
						Max: pulumi.Int(65535),
					},
				},
			},
		},
	})
	if err != nil { return err }

	_, err = rbacV1.NewClusterRole(ctx, "restricted-cluster-role", &rbacV1.ClusterRoleArgs{
		Metadata: metaV1.ObjectMetaArgs{
			Name: pulumi.String("restricted"),
		},
		Rules: rbacV1.PolicyRuleArray{
			rbacV1.PolicyRuleArgs{
				ApiGroups: pulumi.StringArray{
					pulumi.String("policy"),
				},
				ResourceNames: pulumi.StringArray{
					pulumi.String("restricted"),
				},
				Resources: pulumi.StringArray{
					pulumi.String("podsecuritypolicies"),
				},
				Verbs: pulumi.StringArray{
					pulumi.String("use"),
				},
			},
		},
	})
	if err != nil { return err }

	// Create a ClusterRoleBinding for the ServiceAccounts of Namespace kube-system
	// to the ClusterRole that uses the restrictive PodSecurityPolicy.
	_, err = rbacV1.NewClusterRoleBinding(ctx, "allow-restricted-kube-system-crb", &rbacV1.ClusterRoleBindingArgs{
		Metadata: metaV1.ObjectMetaArgs{
			Name: pulumi.String("allow-restricted-kube-system"),
		},
		RoleRef: rbacV1.RoleRefArgs{
			ApiGroup: pulumi.String("rbac.authorization.k8s.io"),
			Kind: pulumi.String("ClusterRole"),
			Name: pulumi.String("restricted"),
		},
		Subjects: rbacV1.SubjectArray{
			rbacV1.SubjectArgs{
				Kind: pulumi.String("Group"),
				Name: pulumi.String("system:serviceaccounts"),
				Namespace: pulumi.String("kube-system"),
			},
		},
	})

	// Create a ClusterRoleBinding for the RBAC group automation
	// to the ClusterRole that uses the restrictive PodSecurityPolicy.
	_, err = rbacV1.NewClusterRoleBinding(ctx, "allow-restricted-apps-crb", &rbacV1.ClusterRoleBindingArgs{
		Metadata: metaV1.ObjectMetaArgs{
			Name: pulumi.String("allow-restricted-apps"),
		},
		RoleRef: rbacV1.RoleRefArgs{
			ApiGroup: pulumi.String("rbac.authorization.k8s.io"),
			Kind: pulumi.String("ClusterRole"),
			Name: pulumi.String("restricted"),
		},
		Subjects: rbacV1.SubjectArray{
			rbacV1.SubjectArgs{
				Kind: pulumi.String("Group"),
				Name: pulumi.String("automation"),
			},
		},
	})
	return nil
}

func ConfigureAutomationRole(ctx *pulumi.Context) error {
	_, err := rbacV1.NewRole(ctx, "k8s-automation-role", &rbacV1.RoleArgs{
		Metadata: metaV1.ObjectMetaArgs{
			Namespace: pulumi.String("apps"),
			Name: pulumi.String("automation-role"),
		},
		Rules: rbacV1.PolicyRuleArray{
			rbacV1.PolicyRuleArgs{
				ApiGroups: pulumi.StringArray{
					pulumi.String(""),
				},
				Resources: pulumi.StringArray{
					pulumi.String("pods"),
					pulumi.String("secrets"),
					pulumi.String("services"),
					pulumi.String("persistentvolumeclaims"),
				},
				Verbs: pulumi.StringArray{
					pulumi.String("get"),
					pulumi.String("list"),
					pulumi.String("watch"),
					pulumi.String("create"),
					pulumi.String("update"),
					pulumi.String("delete"),
				},
			},
			rbacV1.PolicyRuleArgs{
				ApiGroups: pulumi.StringArray{
					pulumi.String("extensions"),
					pulumi.String("apps"),
				},
				Resources: pulumi.StringArray{
					pulumi.String("replicasets"),
					pulumi.String("deployments"),
				},
				Verbs: pulumi.StringArray{
					pulumi.String("get"),
					pulumi.String("list"),
					pulumi.String("watch"),
					pulumi.String("create"),
					pulumi.String("update"),
					pulumi.String("delete"),
				},
			},
		},
	})
	if err != nil { return err }

	return nil
}

func ApplyAutomationRoleBindingToNamespace(ctx *pulumi.Context, namespace string) error {
	_, err := rbacV1.NewRoleBinding(ctx, "k8s-automation-role-binding", &rbacV1.RoleBindingArgs{
		Metadata: metaV1.ObjectMetaArgs{
			Namespace: pulumi.String(namespace),
		},
		RoleRef: rbacV1.RoleRefArgs{
			ApiGroup: pulumi.String("rbac.authorization.k8s.io"),
			Kind: pulumi.String("Role"),
			Name: pulumi.String("automation-role"),
		},
		Subjects: rbacV1.SubjectArray{
			rbacV1.SubjectArgs{
				Kind: pulumi.String("Group"),
				Name: pulumi.String("automation"),
			},
		},
	})
	if err != nil { return err }

	return nil
}

func DeployLoggingAndMonitoring(ctx *pulumi.Context, clusterName string, region string) error {
	_, err := coreV1.NewNamespace(ctx, "logging-namespace", &coreV1.NamespaceArgs{
		Metadata: metaV1.ObjectMetaArgs{
			Name: pulumi.String("logging"),
		},
	})
	if err != nil { return err }

	// logging to cloudwatch
	err = DeployFluentbit(ctx, clusterName, region, "logging")
	if err != nil { return nil }

	// cluster metrics to cloudwatch
	err = DeployCloudwatchAgent(ctx, clusterName, region, "logging")
	if err != nil { return nil }

	return nil
}