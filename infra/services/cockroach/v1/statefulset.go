package crdb

import (
	appsv1 "github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes/apps/v1"
	corev1 "github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes/core/v1"
	metav1 "github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes/meta/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type StatefulSetArgs struct {
	Namespace          pulumi.StringPtrInput
	Replicas           int
	CertSecretName     pulumi.StringPtrInput
	ClientSecretName   pulumi.StringPtrInput
	ServiceAccountName pulumi.StringPtrInput
}

func DeployStatefulSet(ctx *pulumi.Context, args StatefulSetArgs, opts ...pulumi.ResourceOption) (*appsv1.StatefulSet, error) {

	var startArgs pulumi.StringInput
	if args.Replicas == 1 {
		startArgs = pulumi.String("exec " +
			"/cockroach/cockroach start-single-node " +
			"--logtostderr " +
			"--certs-dir /cockroach/cockroach-certs " +
			"--advertise-host $(hostname -f) " +
			"--http-addr 0.0.0.0 " +
			"--cache $(expr $MEMORY_LIMIT_MIB / 4)MiB " +
			"--max-sql-memory $(expr $MEMORY_LIMIT_MIB / 4)MiB " +
			"--join cockroachdb-0.cockroachdb,cockroachdb-1.cockroachdb,cockroachdb-2.cockroachdb",
		)
	} else {
		startArgs = pulumi.Sprintf("exec "+
			"/cockroach/cockroach start "+
			"--logtostderr "+
			"--certs-dir /cockroach/cockroach-certs "+
			"--advertise-host $(hostname -f) "+
			"--http-addr 0.0.0.0 "+
			"--cache $(expr $MEMORY_LIMIT_MIB / 4)MiB "+
			"--max-sql-memory $(expr $MEMORY_LIMIT_MIB / 4)MiB "+
			"--join cockroachdb-0.%s,cockroachdb-1.%s,cockroachdb-2.%s", args.Namespace, args.Namespace, args.Namespace)
	}

	return appsv1.NewStatefulSet(ctx, "crdb-statefulSet", &appsv1.StatefulSetArgs{
		ApiVersion: pulumi.String("apps/v1"),
		Kind:       pulumi.String("StatefulSet"),
		Metadata: &metav1.ObjectMetaArgs{
			Name:      pulumi.String("cockroachdb"),
			Namespace: args.Namespace,
		},
		Spec: &appsv1.StatefulSetSpecArgs{
			ServiceName: pulumi.String("cockroachdb"),
			Replicas:    pulumi.Int(args.Replicas),
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
					ServiceAccountName: args.ServiceAccountName,
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
								&corev1.VolumeMountArgs{
									Name:      pulumi.String("client-certs"),
									MountPath: pulumi.String("/cockroach/cockroach-client-certs"),
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
								startArgs,
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
							Projected: &corev1.ProjectedVolumeSourceArgs{
								Sources: &corev1.VolumeProjectionArray{
									&corev1.VolumeProjectionArgs{
										Secret: corev1.SecretProjectionArgs{
											Name: args.CertSecretName,
											Items: corev1.KeyToPathArray{
												corev1.KeyToPathArgs{
													Key:  pulumi.String("ca.crt"),
													Path: pulumi.String("ca.crt"),
													Mode: pulumi.Int(256),
												},
												corev1.KeyToPathArgs{
													Key:  pulumi.String("tls.crt"),
													Path: pulumi.String("node.crt"),
													Mode: pulumi.Int(256),
												},
												corev1.KeyToPathArgs{
													Key:  pulumi.String("tls.key"),
													Path: pulumi.String("node.key"),
													Mode: pulumi.Int(256),
												},
											},
										},
									},
								},
							},
						},
						&corev1.VolumeArgs{
							Name: pulumi.String("client-certs"),
							Projected: &corev1.ProjectedVolumeSourceArgs{
								Sources: &corev1.VolumeProjectionArray{
									&corev1.VolumeProjectionArgs{
										Secret: corev1.SecretProjectionArgs{
											Name: args.ClientSecretName,
											Items: corev1.KeyToPathArray{
												corev1.KeyToPathArgs{
													Key:  pulumi.String("ca.crt"),
													Path: pulumi.String("ca.crt"),
													Mode: pulumi.Int(256),
												},
												corev1.KeyToPathArgs{
													Key:  pulumi.String("tls.crt"),
													Path: pulumi.String("client.root.crt"),
													Mode: pulumi.Int(256),
												},
												corev1.KeyToPathArgs{
													Key:  pulumi.String("tls.key"),
													Path: pulumi.String("client.root.key"),
													Mode: pulumi.Int(256),
												},
											},
										},
									},
								},
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
								"storage": pulumi.String("15Gi"),
							},
						},
						StorageClassName: pulumi.String("ebs-sc"),
					},
				},
			},
		},
	}, opts...)
}
