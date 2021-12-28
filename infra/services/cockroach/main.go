package cockroach

import (
	appsv1 "github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes/apps/v1"
	corev1 "github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes/core/v1"
	metav1 "github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes/meta/v1"
	policyv1beta1 "github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes/policy/v1beta1"
	rbacv1 "github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes/rbac/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func DeployCockroach(ctx *pulumi.Context) error {
	err := rbac(ctx)
	if err != nil {
		return err
	}

	err = statefulSet(ctx)
	if err != nil {
		return err
	}

	err = publicService(ctx)
	if err != nil {
		return err
	}
	err = privateService(ctx)
	if err != nil {
		return err
	}

	err = podDistributionBudget(ctx)
	if err != nil {
		return err
	}

	return nil
}

func podDistributionBudget(ctx *pulumi.Context) error {
	_, err := policyv1beta1.NewPodDisruptionBudget(ctx, "podDisruptionBudget", &policyv1beta1.PodDisruptionBudgetArgs{
		ApiVersion: pulumi.String("policy/v1beta1"),
		Kind:       pulumi.String("PodDisruptionBudget"),
		Metadata: &metav1.ObjectMetaArgs{
			Name: pulumi.String("cockroachdb-budget"),
			Labels: pulumi.StringMap{
				"app": pulumi.String("cockroachdb"),
			},
		},
		Spec: &policyv1beta1.PodDisruptionBudgetSpecArgs{
			Selector: &metav1.LabelSelectorArgs{
				MatchLabels: pulumi.StringMap{
					"app": pulumi.String("cockroachdb"),
				},
			},
			MaxUnavailable: pulumi.Int(1),
		},
	})
	if err != nil {
		return err
	}
	return nil
}

func statefulSet(ctx *pulumi.Context) error {
	_, err := appsv1.NewStatefulSet(ctx, "statefulSet", &appsv1.StatefulSetArgs{
		ApiVersion: pulumi.String("apps/v1"),
		Kind:       pulumi.String("StatefulSet"),
		Metadata: &metav1.ObjectMetaArgs{
			Name: pulumi.String("cockroachdb"),
		},
		Spec: &appsv1.StatefulSetSpecArgs{
			ServiceName: pulumi.String("cockroachdb"),
			Replicas:    pulumi.Int(1),
			Selector: &metav1.LabelSelectorArgs{
				MatchLabels: pulumi.StringMap{
					"app": pulumi.String("cockroachdb"),
				},
			},
			Template: &corev1.PodTemplateSpecArgs{
				Metadata: &metav1.ObjectMetaArgs{
					Labels: pulumi.StringMap{
						"app": pulumi.String("cockroachdb"),
					},
				},
				Spec: &corev1.PodSpecArgs{
					ServiceAccountName: pulumi.String("cockroachdb"),
					Affinity: &corev1.AffinityArgs{
						PodAntiAffinity: &corev1.PodAntiAffinityArgs{
							PreferredDuringSchedulingIgnoredDuringExecution: corev1.WeightedPodAffinityTermArray{
								&corev1.WeightedPodAffinityTermArgs{
									Weight: pulumi.Int(100),
									PodAffinityTerm: &corev1.PodAffinityTermArgs{
										LabelSelector: &metav1.LabelSelectorArgs{
											MatchExpressions: metav1.LabelSelectorRequirementArray{
												&metav1.LabelSelectorRequirementArgs{
													Key:      pulumi.String("app"),
													Operator: pulumi.String("In"),
													Values: pulumi.StringArray{
														pulumi.String("cockroachdb"),
													},
												},
											},
										},
										TopologyKey: pulumi.String("kubernetes.io/hostname"),
									},
								},
							},
						},
					},
					Containers: corev1.ContainerArray{
						&corev1.ContainerArgs{
							Name:            pulumi.String("cockroachdb"),
							Image:           pulumi.String("cockroachdb/cockroach:v21.1.10"),
							ImagePullPolicy: pulumi.String("IfNotPresent"),
							Ports: corev1.ContainerPortArray{
								&corev1.ContainerPortArgs{
									ContainerPort: pulumi.Int(26257),
									Name:          pulumi.String("grpc"),
								},
								&corev1.ContainerPortArgs{
									ContainerPort: pulumi.Int(8080),
									Name:          pulumi.String("http"),
								},
							},
							ReadinessProbe: &corev1.ProbeArgs{
								HttpGet: &corev1.HTTPGetActionArgs{
									Path:   pulumi.String("/health?ready=1"),
									Port:   pulumi.String("http"),
									Scheme: pulumi.String("HTTPS"),
								},
								InitialDelaySeconds: pulumi.Int(10),
								PeriodSeconds:       pulumi.Int(5),
								FailureThreshold:    pulumi.Int(2),
							},
							VolumeMounts: corev1.VolumeMountArray{
								&corev1.VolumeMountArgs{
									Name:      pulumi.String("datadir"),
									MountPath: pulumi.String("/cockroach/cockroach-data"),
								},
								&corev1.VolumeMountArgs{
									Name:      pulumi.String("certs"),
									MountPath: pulumi.String("/cockroach/cockroach-certs"),
								},
							},
							Env: corev1.EnvVarArray{
								&corev1.EnvVarArgs{
									Name:  pulumi.String("COCKROACH_CHANNEL"),
									Value: pulumi.String("kubernetes-secure"),
								},
								&corev1.EnvVarArgs{
									Name: pulumi.String("GOMAXPROCS"),
									ValueFrom: &corev1.EnvVarSourceArgs{
										ResourceFieldRef: &corev1.ResourceFieldSelectorArgs{
											Resource: pulumi.String("limits.cpu"),
											Divisor:  pulumi.String("1"),
										},
									},
								},
								&corev1.EnvVarArgs{
									Name: pulumi.String("MEMORY_LIMIT_MIB"),
									ValueFrom: &corev1.EnvVarSourceArgs{
										ResourceFieldRef: &corev1.ResourceFieldSelectorArgs{
											Resource: pulumi.String("limits.memory"),
											Divisor:  pulumi.String("1Mi"),
										},
									},
								},
							},
							Command: pulumi.StringArray{
								pulumi.String("/bin/bash"),
								pulumi.String("-ecx"),
								pulumi.String("exec " +
									"/cockroach/cockroach start-single-node" +
									"--logtostderr " +
									"--certs-dir /cockroach/cockroach-certs" +
									"--advertise-host $(hostname -f)" +
									"--http-addr 0.0.0.0" +
									"--cache $(expr $MEMORY_LIMIT_MIB / 4)MiB" +
									"--max-sql-memory $(expr $MEMORY_LIMIT_MIB / 4)MiB"),
							},
						},
					},
					TerminationGracePeriodSeconds: pulumi.Int(60),
					Volumes: corev1.VolumeArray{
						&corev1.VolumeArgs{
							Name: pulumi.String("datadir"),
							PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSourceArgs{
								ClaimName: pulumi.String("datadir"),
							},
						},
						&corev1.VolumeArgs{
							Name: pulumi.String("certs"),
							Secret: &corev1.SecretVolumeSourceArgs{
								SecretName:  pulumi.String("cockroachdb.node"),
								DefaultMode: pulumi.Int(256),
							},
						},
					},
				},
			},
			PodManagementPolicy: pulumi.String("Parallel"),
			UpdateStrategy: &appsv1.StatefulSetUpdateStrategyArgs{
				Type: pulumi.String("RollingUpdate"),
			},
			VolumeClaimTemplates: corev1.PersistentVolumeClaimTypeArray{
				&corev1.PersistentVolumeClaimTypeArgs{
					Metadata: &metav1.ObjectMetaArgs{
						Name: pulumi.String("datadir"),
					},
					Spec: &corev1.PersistentVolumeClaimSpecArgs{
						AccessModes: pulumi.StringArray{
							pulumi.String("ReadWriteOnce"),
						},
						Resources: &corev1.ResourceRequirementsArgs{
							Requests: pulumi.StringMap{
								"storage": pulumi.String("1Gi"),
							},
						},
					},
				},
			},
		},
	})
	if err != nil {
		return err
	}
	return nil
}

func publicService(ctx *pulumi.Context) error {
	_, err := corev1.NewService(ctx, "public-service", &corev1.ServiceArgs{
		ApiVersion: pulumi.String("v1"),
		Kind:       pulumi.String("Service"),
		Metadata: &metav1.ObjectMetaArgs{
			Name: pulumi.String("cockroachdb-public"),
			Labels: pulumi.StringMap{
				"app": pulumi.String("cockroachdb"),
			},
		},
		Spec: &corev1.ServiceSpecArgs{
			Ports: corev1.ServicePortArray{
				&corev1.ServicePortArgs{
					Port:       pulumi.Int(26257),
					TargetPort: pulumi.Int(26257),
					Name:       pulumi.String("grpc"),
				},
				&corev1.ServicePortArgs{
					Port:       pulumi.Int(8080),
					TargetPort: pulumi.Int(8080),
					Name:       pulumi.String("http"),
				},
			},
			Selector: pulumi.StringMap{
				"app": pulumi.String("cockroachdb"),
			},
		},
	})
	if err != nil {
		return err
	}
	return nil
}

func privateService(ctx *pulumi.Context) error {
	_, err := corev1.NewService(ctx, "private-service", &corev1.ServiceArgs{
		ApiVersion: pulumi.String("v1"),
		Kind:       pulumi.String("Service"),
		Metadata: &metav1.ObjectMetaArgs{
			Name: pulumi.String("cockroachdb"),
			Labels: pulumi.StringMap{
				"app": pulumi.String("cockroachdb"),
			},
			Annotations: pulumi.StringMap{
				"service.alpha.kubernetes.io/tolerate-unready-endpoints": pulumi.String("true"),
				"prometheus.io/scrape": pulumi.String("true"),
				"prometheus.io/path":   pulumi.String("_status/vars"),
				"prometheus.io/port":   pulumi.String("8080"),
			},
		},
		Spec: &corev1.ServiceSpecArgs{
			Ports: corev1.ServicePortArray{
				&corev1.ServicePortArgs{
					Port:       pulumi.Int(26257),
					TargetPort: pulumi.Int(26257),
					Name:       pulumi.String("grpc"),
				},
				&corev1.ServicePortArgs{
					Port:       pulumi.Int(8080),
					TargetPort: pulumi.Int(8080),
					Name:       pulumi.String("http"),
				},
			},
			PublishNotReadyAddresses: pulumi.Bool(true),
			ClusterIP:                pulumi.String("None"),
			Selector: pulumi.StringMap{
				"app": pulumi.String("cockroachdb"),
			},
		},
	})
	if err != nil {
		return err
	}
	return nil
}

func rbac(ctx *pulumi.Context) error {
	_, err := corev1.NewServiceAccount(ctx, "serviceAccount", &corev1.ServiceAccountArgs{
		ApiVersion: pulumi.String("v1"),
		Kind:       pulumi.String("ServiceAccount"),
		Metadata: &metav1.ObjectMetaArgs{
			Name: pulumi.String("cockroachdb"),
			Labels: pulumi.StringMap{
				"app": pulumi.String("cockroachdb"),
			},
		},
		AutomountServiceAccountToken: pulumi.Bool(false),
	})
	if err != nil {
		return err
	}

	_, err = rbacv1.NewRole(ctx, "role", &rbacv1.RoleArgs{
		ApiVersion: pulumi.String("rbac.authorization.k8s.io/v1"),
		Kind:       pulumi.String("Role"),
		Metadata: &metav1.ObjectMetaArgs{
			Name: pulumi.String("cockroachdb"),
			Labels: pulumi.StringMap{
				"app": pulumi.String("cockroachdb"),
			},
		},
		Rules: rbacv1.PolicyRuleArray{
			&rbacv1.PolicyRuleArgs{
				ApiGroups: pulumi.StringArray{
					pulumi.String(""),
				},
				Resources: pulumi.StringArray{
					pulumi.String("secrets"),
				},
				Verbs: pulumi.StringArray{
					pulumi.String("get"),
				},
			},
		},
	})
	if err != nil {
		return err
	}
	_, err = rbacv1.NewRoleBinding(ctx, "role-binding", &rbacv1.RoleBindingArgs{
		ApiVersion: pulumi.String("rbac.authorization.k8s.io/v1"),
		Kind:       pulumi.String("RoleBinding"),
		Metadata: &metav1.ObjectMetaArgs{
			Name: pulumi.String("cockroachdb"),
			Labels: pulumi.StringMap{
				"app": pulumi.String("cockroachdb"),
			},
		},
		RoleRef: &rbacv1.RoleRefArgs{
			ApiGroup: pulumi.String("rbac.authorization.k8s.io"),
			Kind:     pulumi.String("Role"),
			Name:     pulumi.String("cockroachdb"),
		},
		Subjects: rbacv1.SubjectArray{
			&rbacv1.SubjectArgs{
				Kind:      pulumi.String("ServiceAccount"),
				Name:      pulumi.String("cockroachdb"),
				Namespace: pulumi.String("default"),
			},
		},
	})
	if err != nil {
		return err
	}
	return nil
}
