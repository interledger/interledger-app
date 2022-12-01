package kubernetes

import (
	"fmt"

	"github.com/pulumi/pulumi-aws/sdk/v4/go/aws/iam"
	appsV1 "github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes/apps/v1"
	coreV1 "github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes/core/v1"
	metaV1 "github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes/meta/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"gitlab.com/fynbos/infra/aws/modules/utils"
	"gopkg.in/yaml.v3"
)

// This role is able to manage all eks resources and is intended
// to be mapped to the k8s system:masters group which would
// allow full access to all k8s resources.
func NewEksClusterAdminPolicy(ctx *pulumi.Context) string {
	policy, err := iam.GetPolicyDocument(ctx, &iam.GetPolicyDocumentArgs{
		Statements: []iam.GetPolicyDocumentStatement{
			{
				Effect: utils.StringPtr("Allow"),
				Actions: []string{
					"eks:*",
					"ec2:DescribeImages",
				},
				Resources: []string{"*"},
			},
			{
				Effect:    utils.StringPtr("Allow"),
				Actions:   []string{"iam:PassRole"},
				Resources: []string{"*"},
			},
		},
	})
	if err != nil {
		return ""
	}

	return policy.Json
}

func NewEksClusterAdminRoleTrustPolicy(ctx *pulumi.Context, accountId string) string {
	policy, err := iam.GetPolicyDocument(ctx, &iam.GetPolicyDocumentArgs{
		Statements: []iam.GetPolicyDocumentStatement{
			{
				Effect: utils.StringPtr("Allow"),
				Actions: []string{
					"sts:AssumeRole",
				},
				Principals: []iam.GetPolicyDocumentStatementPrincipal{
					{
						Type: "AWS",
						Identifiers: []string{
							//The suffix root in the policy’s Principal attribute equates to
							// “authenticated and authorized principals in the account,”
							// not the special and all-powerful root user principal that is created
							// when an AWS account is created.
							// TODO: this needs to be more tightly scoped once we have more users.  e.g. arn:aws:iam::%s:user/xxx
							fmt.Sprintf("arn:aws:iam::%s:root", accountId),
						},
					},
				},
				Conditions: []iam.GetPolicyDocumentStatementCondition{
					{
						Test:     "Bool",
						Values:   []string{"true"},
						Variable: "aws:MultiFactorAuthPresent",
					},
				},
			},
		},
	})
	if err != nil {
		return ""
	}

	return policy.Json
}

func NewFluentdCloudwatchPolicy(ctx *pulumi.Context) string {
	policy, err := iam.GetPolicyDocument(ctx, &iam.GetPolicyDocumentArgs{
		Statements: []iam.GetPolicyDocumentStatement{
			{
				Effect: utils.StringPtr("Allow"),
				Actions: []string{
					"logs:*",
				},
				Resources: []string{"arn:aws:logs:*:*:*"},
			},
		},
	})
	if err != nil {
		return ""
	}

	return policy.Json
}

func NewEksAutomationRoleTrustPolicy(ctx *pulumi.Context, accountId string, glRunnerRoleArn pulumi.StringOutput) pulumi.StringOutput {
	return pulumi.All(glRunnerRoleArn).ApplyT(func(args []interface{}) (string, error) {
		runnerRole := args[0].(string)

		policy, err := iam.GetPolicyDocument(ctx, &iam.GetPolicyDocumentArgs{
			Statements: []iam.GetPolicyDocumentStatement{
				{
					Effect: utils.StringPtr("Allow"),
					Actions: []string{
						"sts:AssumeRole",
					},
					Principals: []iam.GetPolicyDocumentStatementPrincipal{
						{
							Type:        "AWS",
							Identifiers: []string{runnerRole},
						},
					},
				},
			},
		})
		if err != nil {
			return "", err
		}

		return policy.Json, nil
	}).(pulumi.StringOutput)
}

// Creates a policy definition that will allow access to ALL EKS cluster admin endpoints
// in the specified account. The RBAC for the actual k8s resources will be defined inside
// the cluster through the use of ClusterRoles and ClusterRoleBindings.
func NewEksAutomationRolePolicy(ctx *pulumi.Context, accountId string) (string, error) {
	policy, err := iam.GetPolicyDocument(ctx, &iam.GetPolicyDocumentArgs{
		Statements: []iam.GetPolicyDocumentStatement{
			{
				Effect: utils.StringPtr("Allow"),
				Actions: []string{
					"eks:AccessKubernetesApi",
					"eks:DescribeCluster",
				},
				Resources: []string{"arn:aws:eks:*:" + accountId + ":cluster/*"},
			},
		},
	})
	if err != nil {
		return "", err
	}

	return policy.Json, nil
}

func NewEksNodeGroupRoleTrustPolicy(ctx *pulumi.Context) string {
	policy, err := iam.GetPolicyDocument(ctx, &iam.GetPolicyDocumentArgs{
		Statements: []iam.GetPolicyDocumentStatement{
			{
				Effect: utils.StringPtr("Allow"),
				Actions: []string{
					"sts:AssumeRole",
				},
				Principals: []iam.GetPolicyDocumentStatementPrincipal{
					{
						Type:        "Service",
						Identifiers: []string{"ec2.amazonaws.com"},
					},
				},
			},
		},
	})
	if err != nil {
		return ""
	}

	return policy.Json
}

type EksIamRoles struct {
	Admin      *iam.Role
	Automation *iam.Role
	NodeGroup  *iam.Role
}

func NewEksRoles(ctx *pulumi.Context, accountId string, glRunnerRoleArn pulumi.StringOutput) (EksIamRoles, error) {
	var ret EksIamRoles
	adminRole, err := iam.NewRole(ctx, "eks-cluster-admin-role", &iam.RoleArgs{
		Name:             pulumi.String("eksClusterAdminRole"),
		Description:      pulumi.String("Admin role for eks cluster"),
		AssumeRolePolicy: pulumi.String(NewEksClusterAdminRoleTrustPolicy(ctx, accountId)),
		InlinePolicies: iam.RoleInlinePolicyArray{
			iam.RoleInlinePolicyArgs{
				Name:   pulumi.String("eks-admin-policy"),
				Policy: pulumi.String(NewEksClusterAdminPolicy(ctx)),
			},
		},
	})
	if err != nil {
		return ret, err
	}

	automationPolicy, err := NewEksAutomationRolePolicy(ctx, accountId)
	if err != nil {
		return ret, err
	}
	automationRole, err := iam.NewRole(ctx, "eks-automation-role", &iam.RoleArgs{
		Name:             pulumi.String("eksAutomationRole"),
		Description:      pulumi.String("Automation role for eks cluster"),
		AssumeRolePolicy: NewEksAutomationRoleTrustPolicy(ctx, accountId, glRunnerRoleArn),
		InlinePolicies: iam.RoleInlinePolicyArray{
			iam.RoleInlinePolicyArgs{
				Name:   pulumi.String("eksReadClusters"),
				Policy: pulumi.String(automationPolicy),
			},
		},
	})
	if err != nil {
		return ret, err
	}

	nodeGroupRole, err := iam.NewRole(ctx, "eks-cluster-node-group-role", &iam.RoleArgs{
		Name:             pulumi.String("eksNodeGroupRole"),
		Description:      pulumi.String("Node group role for eks cluster"),
		AssumeRolePolicy: pulumi.String(NewEksNodeGroupRoleTrustPolicy(ctx)),
		ManagedPolicyArns: pulumi.StringArray{
			pulumi.String("arn:aws:iam::aws:policy/AmazonEKSWorkerNodePolicy"),
			pulumi.String("arn:aws:iam::aws:policy/AmazonEKS_CNI_Policy"),
			pulumi.String("arn:aws:iam::aws:policy/AmazonEC2ContainerRegistryReadOnly"),
		},
		InlinePolicies: iam.RoleInlinePolicyArray{
			iam.RoleInlinePolicyArgs{
				Name:   pulumi.String("fluentd-cloudwatch-policy"),
				Policy: pulumi.String(NewFluentdCloudwatchPolicy(ctx)),
			},
		},
	})
	if err != nil {
		return ret, err
	}

	ret.Admin = adminRole
	ret.Automation = automationRole
	ret.NodeGroup = nodeGroupRole
	return ret, nil
}

// This follows the EKS best practices of enabling IRSA for aws-node by:
// 1. Updating the aws-node cni service account to use an AWS IAM role that has the aws managed EKS_CNI policy.
// 2. Enable IRSA on the aws-node daemon-set.
//
// Note that the aws-node service account and aws-node daemon-set first have to be imported using the pulumi cli as it is created by EKS.
// See readme on how to do this. Note that these resources must therefore remain protected.
func UpdateAwsNodeDaemonSetToUseIrsa(ctx *pulumi.Context, accountId string, oidcProvider string) error {
	trustPolicy, err := NewIamTrustPolicyDocument(ctx, accountId, oidcProvider, "kube-system", "aws-node")
	if err != nil {
		return err
	}
	role, err := iam.NewRole(ctx, "eks-cluster-aws-node-role", &iam.RoleArgs{
		Name:             pulumi.String("eksAwsNodeRole"),
		Description:      pulumi.String("Role for aws-node daemon set"),
		AssumeRolePolicy: pulumi.String(trustPolicy),
		ManagedPolicyArns: pulumi.StringArray{
			pulumi.String("arn:aws:iam::aws:policy/AmazonEKS_CNI_Policy"),
		},
	})
	if err != nil {
		return err
	}

	// This resource s deployed by EKS and has to be imported using the pulumi cli. See readme on how to do this.
	// It is critical that this resource remain protected as it wasn't deployed by pulumi.
	_, err = coreV1.NewServiceAccount(ctx, "aws-node-sa", &coreV1.ServiceAccountArgs{
		ApiVersion: pulumi.String("v1"),
		Kind:       pulumi.String("ServiceAccount"),
		Metadata: metaV1.ObjectMetaArgs{
			Annotations: pulumi.StringMap{
				"eks.amazonaws.com/role-arn": role.Arn,
			},
			Name:      pulumi.String("aws-node"),
			Namespace: pulumi.String("kube-system"),
		},
	}, pulumi.Protect(true)) // NB: leave protected. see above comment.
	if err != nil {
		return err
	}

	// This resource s deployed by EKS and has to be imported using the pulumi cli. See readme on how to do this.
	// It is critical that this resource remain protected as it wasn't deployed by pulumi.
	_, err = appsV1.NewDaemonSet(ctx, "aws-node-ds", &appsV1.DaemonSetArgs{
		ApiVersion: pulumi.String("apps/v1"),
		Kind:       pulumi.String("DaemonSet"),
		Metadata: &metaV1.ObjectMetaArgs{
			Annotations: pulumi.StringMap{
				"irsa": pulumi.String("enabled"),
			},
			Labels: pulumi.StringMap{
				"k8s-app": pulumi.String("aws-node"),
			},
			Name:      pulumi.String("aws-node"),
			Namespace: pulumi.String("kube-system"),
		},
		Spec: &appsV1.DaemonSetSpecArgs{
			Selector: &metaV1.LabelSelectorArgs{
				MatchLabels: pulumi.StringMap{
					"k8s-app": pulumi.String("aws-node"),
				},
			},
			Template: &coreV1.PodTemplateSpecArgs{
				Metadata: &metaV1.ObjectMetaArgs{
					CreationTimestamp: nil,
					Labels: pulumi.StringMap{
						"k8s-app": pulumi.String("aws-node"),
					},
				},
				Spec: &coreV1.PodSpecArgs{
					Affinity: &coreV1.AffinityArgs{
						NodeAffinity: &coreV1.NodeAffinityArgs{
							RequiredDuringSchedulingIgnoredDuringExecution: &coreV1.NodeSelectorArgs{
								NodeSelectorTerms: coreV1.NodeSelectorTermArray{
									&coreV1.NodeSelectorTermArgs{
										MatchExpressions: coreV1.NodeSelectorRequirementArray{
											&coreV1.NodeSelectorRequirementArgs{
												Key:      pulumi.String("beta.kubernetes.io/os"),
												Operator: pulumi.String("In"),
												Values: pulumi.StringArray{
													pulumi.String("linux"),
												},
											},
											&coreV1.NodeSelectorRequirementArgs{
												Key:      pulumi.String("beta.kubernetes.io/arch"),
												Operator: pulumi.String("In"),
												Values: pulumi.StringArray{
													pulumi.String("amd64"),
													pulumi.String("arm64"),
												},
											},
											&coreV1.NodeSelectorRequirementArgs{
												Key:      pulumi.String("eks.amazonaws.com/compute-type"),
												Operator: pulumi.String("NotIn"),
												Values: pulumi.StringArray{
													pulumi.String("fargate"),
												},
											},
										},
									},
									&coreV1.NodeSelectorTermArgs{
										MatchExpressions: coreV1.NodeSelectorRequirementArray{
											&coreV1.NodeSelectorRequirementArgs{
												Key:      pulumi.String("kubernetes.io/os"),
												Operator: pulumi.String("In"),
												Values: pulumi.StringArray{
													pulumi.String("linux"),
												},
											},
											&coreV1.NodeSelectorRequirementArgs{
												Key:      pulumi.String("kubernetes.io/arch"),
												Operator: pulumi.String("In"),
												Values: pulumi.StringArray{
													pulumi.String("amd64"),
													pulumi.String("arm64"),
												},
											},
											&coreV1.NodeSelectorRequirementArgs{
												Key:      pulumi.String("eks.amazonaws.com/compute-type"),
												Operator: pulumi.String("NotIn"),
												Values: pulumi.StringArray{
													pulumi.String("fargate"),
												},
											},
										},
									},
								},
							},
						},
					},
					Containers: coreV1.ContainerArray{
						&coreV1.ContainerArgs{
							Env: coreV1.EnvVarArray{
								&coreV1.EnvVarArgs{
									Name:  pulumi.String("AWS_VPC_CNI_NODE_PORT_SUPPORT"),
									Value: pulumi.String("true"),
								},
								&coreV1.EnvVarArgs{
									Name:  pulumi.String("AWS_VPC_ENI_MTU"),
									Value: pulumi.String("9001"),
								},
								&coreV1.EnvVarArgs{
									Name:  pulumi.String("AWS_VPC_K8S_CNI_CONFIGURE_RPFILTER"),
									Value: pulumi.String("false"),
								},
								&coreV1.EnvVarArgs{
									Name:  pulumi.String("AWS_VPC_K8S_CNI_CUSTOM_NETWORK_CFG"),
									Value: pulumi.String("false"),
								},
								&coreV1.EnvVarArgs{
									Name:  pulumi.String("AWS_VPC_K8S_CNI_EXTERNALSNAT"),
									Value: pulumi.String("false"),
								},
								&coreV1.EnvVarArgs{
									Name:  pulumi.String("AWS_VPC_K8S_CNI_LOGLEVEL"),
									Value: pulumi.String("DEBUG"),
								},
								&coreV1.EnvVarArgs{
									Name:  pulumi.String("AWS_VPC_K8S_CNI_LOG_FILE"),
									Value: pulumi.String("/host/var/log/aws-routed-eni/ipamd.log"),
								},
								&coreV1.EnvVarArgs{
									Name:  pulumi.String("AWS_VPC_K8S_CNI_RANDOMIZESNAT"),
									Value: pulumi.String("prng"),
								},
								&coreV1.EnvVarArgs{
									Name:  pulumi.String("AWS_VPC_K8S_CNI_VETHPREFIX"),
									Value: pulumi.String("eni"),
								},
								&coreV1.EnvVarArgs{
									Name:  pulumi.String("AWS_VPC_K8S_PLUGIN_LOG_FILE"),
									Value: pulumi.String("/var/log/aws-routed-eni/plugin.log"),
								},
								&coreV1.EnvVarArgs{
									Name:  pulumi.String("AWS_VPC_K8S_PLUGIN_LOG_LEVEL"),
									Value: pulumi.String("DEBUG"),
								},
								&coreV1.EnvVarArgs{
									Name:  pulumi.String("DISABLE_INTROSPECTION"),
									Value: pulumi.String("false"),
								},
								&coreV1.EnvVarArgs{
									Name:  pulumi.String("DISABLE_METRICS"),
									Value: pulumi.String("false"),
								},
								&coreV1.EnvVarArgs{
									Name:  pulumi.String("DISABLE_NETWORK_RESOURCE_PROVISIONING"),
									Value: pulumi.String("false"),
								},
								&coreV1.EnvVarArgs{
									Name:  pulumi.String("ENABLE_POD_ENI"),
									Value: pulumi.String("false"),
								},
								&coreV1.EnvVarArgs{
									Name:  pulumi.String("ENABLE_PREFIX_DELEGATION"),
									Value: pulumi.String("false"),
								},
								&coreV1.EnvVarArgs{
									Name: pulumi.String("MY_NODE_NAME"),
									ValueFrom: &coreV1.EnvVarSourceArgs{
										FieldRef: &coreV1.ObjectFieldSelectorArgs{
											FieldPath: pulumi.String("spec.nodeName"),
										},
									},
								},
								&coreV1.EnvVarArgs{
									Name:  pulumi.String("WARM_ENI_TARGET"),
									Value: pulumi.String("1"),
								},
								&coreV1.EnvVarArgs{
									Name:  pulumi.String("WARM_PREFIX_TARGET"),
									Value: pulumi.String("1"),
								},
							},
							Image:           pulumi.String("602401143452.dkr.ecr.us-west-2.amazonaws.com/amazon-k8s-cni:v1.9.3"),
							ImagePullPolicy: pulumi.String("Always"),
							LivenessProbe: &coreV1.ProbeArgs{
								Exec: &coreV1.ExecActionArgs{
									Command: pulumi.StringArray{
										pulumi.String("/app/grpc-health-probe"),
										pulumi.String("-addr=:50051"),
										pulumi.String("-connect-timeout=2s"),
										pulumi.String("-rpc-timeout=2s"),
									},
								},
								InitialDelaySeconds: pulumi.Int(60),
							},
							Name: pulumi.String("aws-node"),
							Ports: coreV1.ContainerPortArray{
								&coreV1.ContainerPortArgs{
									ContainerPort: pulumi.Int(61678),
									Name:          pulumi.String("metrics"),
								},
							},
							ReadinessProbe: &coreV1.ProbeArgs{
								Exec: &coreV1.ExecActionArgs{
									Command: pulumi.StringArray{
										pulumi.String("/app/grpc-health-probe"),
										pulumi.String("-addr=:50051"),
										pulumi.String("-connect-timeout=2s"),
										pulumi.String("-rpc-timeout=2s"),
									},
								},
								InitialDelaySeconds: pulumi.Int(1),
							},
							Resources: &coreV1.ResourceRequirementsArgs{
								Requests: pulumi.StringMap{
									"cpu": pulumi.String("10m"),
								},
							},
							SecurityContext: &coreV1.SecurityContextArgs{
								Capabilities: &coreV1.CapabilitiesArgs{
									Add: pulumi.StringArray{
										pulumi.String("NET_ADMIN"),
									},
								},
							},
							VolumeMounts: coreV1.VolumeMountArray{
								&coreV1.VolumeMountArgs{
									MountPath: pulumi.String("/host/opt/cni/bin"),
									Name:      pulumi.String("cni-bin-dir"),
								},
								&coreV1.VolumeMountArgs{
									MountPath: pulumi.String("/host/etc/cni/net.d"),
									Name:      pulumi.String("cni-net-dir"),
								},
								&coreV1.VolumeMountArgs{
									MountPath: pulumi.String("/host/var/log/aws-routed-eni"),
									Name:      pulumi.String("log-dir"),
								},
								&coreV1.VolumeMountArgs{
									MountPath: pulumi.String("/var/run/aws-node"),
									Name:      pulumi.String("run-dir"),
								},
								&coreV1.VolumeMountArgs{
									MountPath: pulumi.String("/var/run/dockershim.sock"),
									Name:      pulumi.String("dockershim"),
								},
								&coreV1.VolumeMountArgs{
									MountPath: pulumi.String("/run/xtables.lock"),
									Name:      pulumi.String("xtables-lock"),
								},
							},
						},
					},
					HostNetwork: pulumi.Bool(true),
					InitContainers: coreV1.ContainerArray{
						&coreV1.ContainerArgs{
							Env: coreV1.EnvVarArray{
								&coreV1.EnvVarArgs{
									Name:  pulumi.String("DISABLE_TCP_EARLY_DEMUX"),
									Value: pulumi.String("false"),
								},
							},
							Image:           pulumi.String("602401143452.dkr.ecr.us-west-2.amazonaws.com/amazon-k8s-cni-init:v1.9.3"),
							ImagePullPolicy: pulumi.String("Always"),
							Name:            pulumi.String("aws-vpc-cni-init"),
							Resources:       nil,
							SecurityContext: &coreV1.SecurityContextArgs{
								Privileged: pulumi.Bool(true),
							},
							VolumeMounts: coreV1.VolumeMountArray{
								&coreV1.VolumeMountArgs{
									MountPath: pulumi.String("/host/opt/cni/bin"),
									Name:      pulumi.String("cni-bin-dir"),
								},
							},
						},
					},
					PriorityClassName:             pulumi.String("system-node-critical"),
					SecurityContext:               nil,
					ServiceAccountName:            pulumi.String("aws-node"),
					TerminationGracePeriodSeconds: pulumi.Int(10),
					Tolerations: coreV1.TolerationArray{
						&coreV1.TolerationArgs{
							Operator: pulumi.String("Exists"),
						},
					},
					Volumes: coreV1.VolumeArray{
						&coreV1.VolumeArgs{
							HostPath: &coreV1.HostPathVolumeSourceArgs{
								Path: pulumi.String("/opt/cni/bin"),
							},
							Name: pulumi.String("cni-bin-dir"),
						},
						&coreV1.VolumeArgs{
							HostPath: &coreV1.HostPathVolumeSourceArgs{
								Path: pulumi.String("/etc/cni/net.d"),
							},
							Name: pulumi.String("cni-net-dir"),
						},
						&coreV1.VolumeArgs{
							HostPath: &coreV1.HostPathVolumeSourceArgs{
								Path: pulumi.String("/var/run/dockershim.sock"),
							},
							Name: pulumi.String("dockershim"),
						},
						&coreV1.VolumeArgs{
							HostPath: &coreV1.HostPathVolumeSourceArgs{
								Path: pulumi.String("/run/xtables.lock"),
							},
							Name: pulumi.String("xtables-lock"),
						},
						&coreV1.VolumeArgs{
							HostPath: &coreV1.HostPathVolumeSourceArgs{
								Path: pulumi.String("/var/log/aws-routed-eni"),
								Type: pulumi.String("DirectoryOrCreate"),
							},
							Name: pulumi.String("log-dir"),
						},
						&coreV1.VolumeArgs{
							HostPath: &coreV1.HostPathVolumeSourceArgs{
								Path: pulumi.String("/var/run/aws-node"),
								Type: pulumi.String("DirectoryOrCreate"),
							},
							Name: pulumi.String("run-dir"),
						},
					},
				},
			},
			UpdateStrategy: &appsV1.DaemonSetUpdateStrategyArgs{
				RollingUpdate: &appsV1.RollingUpdateDaemonSetArgs{
					MaxUnavailable: pulumi.String(fmt.Sprintf("%v%v", "10", "%")),
				},
				Type: pulumi.String("RollingUpdate"),
			},
		},
	}, pulumi.Import(pulumi.ID("kube-system/aws-node")), pulumi.Protect(true))
	if err != nil {
		return err
	}

	return nil
}

func NewIamTrustPolicyDocument(ctx *pulumi.Context, accountId string, oidcProvider string, namespace string, serviceAccount string) (string, error) {

	conditions := []iam.GetPolicyDocumentStatementCondition{
		{
			Test:     "StringEquals",
			Values:   []string{"sts.amazonaws.com"},
			Variable: fmt.Sprintf("%s:aud", oidcProvider),
		},
	}

	if namespace != "" && serviceAccount != "" {
		conditions = append(conditions, iam.GetPolicyDocumentStatementCondition{
			Test: "StringEquals",
			Values: []string{
				fmt.Sprintf("system:serviceaccount:%s:%s", namespace, serviceAccount),
			},
			Variable: fmt.Sprintf("%s:sub", oidcProvider),
		})
	}

	policy, err := iam.GetPolicyDocument(ctx, &iam.GetPolicyDocumentArgs{
		Statements: []iam.GetPolicyDocumentStatement{
			{
				Effect: utils.StringPtr("Allow"),
				Actions: []string{
					"sts:AssumeRoleWithWebIdentity",
				},
				Principals: []iam.GetPolicyDocumentStatementPrincipal{
					{
						Type: "Federated",
						Identifiers: []string{
							fmt.Sprintf("arn:aws:iam::%s:oidc-provider/%s", accountId, oidcProvider),
						},
					},
				},
				Conditions: conditions,
			},
		},
	})
	if err != nil {
		return "", err
	}

	return policy.Json, nil
}

func NewIamTrustPolicyDocumentV2(ctx *pulumi.Context, accountId pulumi.StringPtrInput, oidcProvider pulumi.StringPtrInput, namespace pulumi.StringPtrInput, serviceAccount pulumi.StringPtrInput) pulumi.StringOutput {
	return pulumi.All(accountId, oidcProvider, namespace, serviceAccount).ApplyT(func(args []interface{}) (string, error) {
		aid := args[0].(string)
		oidc := args[1].(string)
		ns := args[2].(string)
		sa := args[3].(string)

		conditions := []iam.GetPolicyDocumentStatementCondition{
			{
				Test:     "StringEquals",
				Values:   []string{"sts.amazonaws.com"},
				Variable: fmt.Sprintf("%s:aud", oidc),
			},
		}

		if ns != "" && sa != "" {
			conditions = append(conditions, iam.GetPolicyDocumentStatementCondition{
				Test: "StringEquals",
				Values: []string{
					fmt.Sprintf("system:serviceaccount:%s:%s", ns, sa),
				},
				Variable: fmt.Sprintf("%s:sub", oidc),
			})
		}

		policy, err := iam.GetPolicyDocument(ctx, &iam.GetPolicyDocumentArgs{
			Statements: []iam.GetPolicyDocumentStatement{
				{
					Effect: utils.StringPtr("Allow"),
					Actions: []string{
						"sts:AssumeRoleWithWebIdentity",
					},
					Principals: []iam.GetPolicyDocumentStatementPrincipal{
						{
							Type: "Federated",
							Identifiers: []string{
								fmt.Sprintf("arn:aws:iam::%s:oidc-provider/%s", aid, oidc),
							},
						},
					},
					Conditions: conditions,
				},
			},
		})
		if err != nil {
			return "", err
		}

		return policy.Json, nil
	}).(pulumi.StringOutput)
}

type ServiceRoleArgs struct {
	Service           string
	Description       string
	ManagedPolicyArns []string
}

func NewServiceRole(ctx *pulumi.Context, name string, args *ServiceRoleArgs) (*iam.Role, error) {

	assumePolicy, err := iam.GetPolicyDocument(ctx, &iam.GetPolicyDocumentArgs{
		Statements: []iam.GetPolicyDocumentStatement{
			{
				Effect: utils.StringPtr("Allow"),
				Actions: []string{
					"sts:AssumeRole",
				},
				Principals: []iam.GetPolicyDocumentStatementPrincipal{
					{
						Type: "Service",
						Identifiers: []string{
							args.Service,
						},
					},
				},
			},
		},
	})
	if err != nil {
		return nil, err
	}

	role, err := iam.NewRole(ctx, name, &iam.RoleArgs{
		Description:       pulumi.String(args.Description),
		AssumeRolePolicy:  pulumi.String(assumePolicy.Json),
		ManagedPolicyArns: pulumi.ToStringArray(args.ManagedPolicyArns),
	})

	return role, err
}

type RoleMap struct {
	RoleArn  string   `yaml:"rolearn"`
	Username string   `yaml:"username"`
	Groups   []string `yaml:"groups"`
}

type ConfigData struct {
	MapRoles string `yaml:"mapRoles"`
}

func RoleMappingConfig(instanceRoles []*iam.Role, mappings []RoleMap) pulumi.StringOutput {
	var inputs []interface{}
	for _, a := range instanceRoles {
		inputs = append(inputs, a.Arn)
	}

	output := pulumi.All(inputs...).ApplyT(func(args []interface{}) (string, error) {
		var roles []RoleMap

		for i := 0; i < len(args); i++ {
			roleArn := args[i].(string)
			r := RoleMap{
				RoleArn:  roleArn,
				Username: "system:node:{{EC2PrivateDNSName}}",
				Groups: []string{
					"system:bootstrappers",
					"system:nodes",
				},
			}
			roles = append(roles, r)
		}

		for _, m := range mappings {
			roles = append(roles, m)
		}

		srm, err := yaml.Marshal(&roles)
		if err != nil {
			return "", err
		}

		return string(srm), nil
	}).(pulumi.StringOutput)

	return output
}

type RoleMapArg struct {
	RoleArn  pulumi.StringInput
	Username pulumi.StringInput
	Groups   pulumi.StringArrayInput
}

// RoleMappingConfigV2 Allows us to define the role map using pulumi inputs, and it will generate the yaml
func RoleMappingConfigV2(mappings []RoleMapArg) pulumi.StringOutput {

	var roleMapping []pulumi.StringOutput

	for _, r := range mappings {
		output := pulumi.All(r.RoleArn, r.Username, r.Groups).ApplyT(func(args []interface{}) (string, error) {
			role := RoleMap{
				RoleArn:  args[0].(string),
				Username: args[1].(string),
				Groups:   args[2].([]string),
			}
			srm, err := yaml.Marshal(role)
			if err != nil {
				return "", err
			}

			return string(srm), nil
		}).(pulumi.StringOutput)
		roleMapping = append(roleMapping, output)
	}

	var inputs []interface{}
	for _, a := range roleMapping {
		inputs = append(inputs, a)
	}

	output := pulumi.All(inputs...).ApplyT(func(args []interface{}) (string, error) {
		var roles []RoleMap

		for _, a := range args {
			var role RoleMap
			err := yaml.Unmarshal([]byte(a.(string)), &role)
			roles = append(roles, role)
			if err != nil {
				return "", err
			}
		}

		srm, err := yaml.Marshal(roles)

		return string(srm), err
	}).(pulumi.StringOutput)

	return output
}

func NewDeployRole(ctx *pulumi.Context, clusterName string, deployerRoleArn string) (*iam.Role, error) {

	trustPolicy, err := iam.GetPolicyDocument(ctx, &iam.GetPolicyDocumentArgs{
		Statements: []iam.GetPolicyDocumentStatement{
			{
				Effect: utils.StringPtr("Allow"),
				Actions: []string{
					"sts:AssumeRole",
				},
				Principals: []iam.GetPolicyDocumentStatementPrincipal{
					{
						Type: "AWS",
						Identifiers: []string{
							deployerRoleArn,
						},
					},
				},
			},
		},
	})
	if err != nil {
		return nil, err
	}

	deployerRole, err := iam.NewRole(ctx, "deploy-role", &iam.RoleArgs{
		Name:             pulumi.String(fmt.Sprintf("eks-%s-deployer", clusterName)),
		AssumeRolePolicy: pulumi.String(trustPolicy.Json),
	})
	if err != nil {
		return nil, err
	}

	return deployerRole, nil
}

func NewBucketReadWriteDeleteAccessPolicy(ctx *pulumi.Context, bucketARN string) (*iam.GetPolicyDocumentResult, error) {
	readWritePolicy, err := iam.GetPolicyDocument(ctx, &iam.GetPolicyDocumentArgs{
		Version: utils.StringPtr("2012-10-17"),
		Statements: []iam.GetPolicyDocumentStatement{
			{
				Sid:       utils.StringPtr("ListObjectsInBucket"),
				Effect:    utils.StringPtr("Allow"),
				Actions:   []string{"s3:ListBucket"},
				Resources: []string{bucketARN},
			},
			{
				Sid:    utils.StringPtr("AllObjectActions"),
				Effect: utils.StringPtr("Allow"),
				Actions: []string{
					"s3:PutObject",
					"s3:GetObject",
					"s3:DeleteObject",
				},
				Resources: []string{fmt.Sprintf("%s/*", bucketARN)},
			},
		},
	})
	if err != nil {
		return nil, err
	}

	return readWritePolicy, nil
}
