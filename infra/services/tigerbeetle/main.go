package tigerbeetle

import (
	appsv1 "github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes/apps/v1"
	corev1 "github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes/core/v1"
	metav1 "github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes/meta/v1"
	rbacv1 "github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes/rbac/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type DeployTigerBeetleArgs struct {
	IsLocal bool
}

func DeployTigerBeetle(ctx *pulumi.Context, args DeployTigerBeetleArgs) error {

	err := rbac(ctx)
	if err != nil {
		return err
	}

	err = statefulSet(ctx, args.IsLocal)
	if err != nil {
		return err
	}

	err = service(ctx)
	if err != nil {
		return err
	}

	return nil
}

func statefulSet(ctx *pulumi.Context, isLocal bool) error {
	_, err := appsv1.NewStatefulSet(ctx, "tigerbeetle-ss", &appsv1.StatefulSetArgs{
		ApiVersion: pulumi.String("apps/v1"),
		Kind:       pulumi.String("StatefulSet"),
		Metadata: &metav1.ObjectMetaArgs{
			Name: pulumi.String("tigerbeetle"),
		},
		Spec: &appsv1.StatefulSetSpecArgs{
			Selector: &metav1.LabelSelectorArgs{
				MatchLabels: pulumi.StringMap{
					"app": pulumi.String("tigerbeetle"),
				},
			},
			ServiceName: pulumi.String("tigerbeetle"),
			Replicas:    pulumi.Int(1),
			Template: &corev1.PodTemplateSpecArgs{
				Metadata: &metav1.ObjectMetaArgs{
					Labels: pulumi.StringMap{
						"app": pulumi.String("tigerbeetle"),
					},
				},
				Spec: &corev1.PodSpecArgs{
					ServiceAccountName: pulumi.String("tigerbeetle"),
					Affinity: &corev1.AffinityArgs{
						PodAntiAffinity: &corev1.PodAntiAffinityArgs{
							RequiredDuringSchedulingIgnoredDuringExecution: corev1.PodAffinityTermArray{
								&corev1.PodAffinityTermArgs{
									LabelSelector: &metav1.LabelSelectorArgs{
										MatchExpressions: metav1.LabelSelectorRequirementArray{
											&metav1.LabelSelectorRequirementArgs{
												Key:      pulumi.String("app"),
												Operator: pulumi.String("In"),
												Values: pulumi.StringArray{
													pulumi.String("tigerbeetle"),
												},
											},
										},
									},
									TopologyKey: pulumi.String("kubernetes.io/hostname"),
								},
							},
						},
					},
					TerminationGracePeriodSeconds: pulumi.Int(10),
					InitContainers: corev1.ContainerArray{
						&corev1.ContainerArgs{
							Name:  pulumi.String("init-tigerbeetle"),
							Image: pulumi.String("donchangfoot/tigerbeetle"),
							VolumeMounts: corev1.VolumeMountArray{
								&corev1.VolumeMountArgs{
									Name:      pulumi.String("datadir"),
									MountPath: pulumi.String("/var/lib/tigerbeetle"),
								},
							},
							Env: corev1.EnvVarArray{
								&corev1.EnvVarArgs{
									Name:  pulumi.String("CLUSTER"),
									Value: pulumi.String("0"),
								},
							},
							Command: pulumi.StringArray{
								pulumi.String("/bin/bash"),
								pulumi.String("-c"),
								pulumi.String("set -ex; REPLICA=${HOSTNAME##*-}; FMT_CLUSTER=$(echo $CLUSTER | awk '{ printf(\"%010d\\n\", $1) }'); FMT_REPLICA=$(echo $REPLICA | awk '{ printf(\"%03d\\n\", $1) }'); ls /var/lib/tigerbeetle; DATA_FILE=/var/lib/tigerbeetle/cluster_${FMT_CLUSTER}_replica_${FMT_REPLICA}.tigerbeetle; if [[ ! -f \"$DATA_FILE\" ]]; then /opt/beta-beetle/tigerbeetle init --cluster=$CLUSTER --replica=$REPLICA --directory=/var/lib/tigerbeetle; fi"),
							},
							// Required for Docker on Mac
							SecurityContext: &corev1.SecurityContextArgs{
								Privileged: pulumi.Bool(isLocal),
							},
						},
					},
					Containers: corev1.ContainerArray{
						&corev1.ContainerArgs{
							Name:  pulumi.String("tigerbeetle"),
							Image: pulumi.String("donchangfoot/tigerbeetle"),
							Ports: corev1.ContainerPortArray{
								&corev1.ContainerPortArgs{
									ContainerPort: pulumi.Int(8080),
									Name:          pulumi.String("http"),
								},
							},
							VolumeMounts: corev1.VolumeMountArray{
								&corev1.VolumeMountArgs{
									Name:      pulumi.String("datadir"),
									MountPath: pulumi.String("/var/lib/tigerbeetle"),
								},
							},
							Env: corev1.EnvVarArray{
								&corev1.EnvVarArgs{
									Name:  pulumi.String("CLUSTER"),
									Value: pulumi.String("0"),
								},
							},
							Command: pulumi.StringArray{
								pulumi.String("/bin/bash"),
								pulumi.String("-c"),
								pulumi.String("set -ex; REPLICA=${HOSTNAME##*-}; /opt/beta-beetle/tigerbeetle start --cluster=$CLUSTER --replica=$REPLICA --directory=/var/lib/tigerbeetle --addresses=0.0.0.0:8080"),
							},
							// Required for Docker on Mac
							SecurityContext: &corev1.SecurityContextArgs{
								Privileged: pulumi.Bool(isLocal),
							},
						},
					},
					Volumes: corev1.VolumeArray{
						&corev1.VolumeArgs{
							Name: pulumi.String("datadir"),
							PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSourceArgs{
								ClaimName: pulumi.String("datadir"),
							},
						},
					},
				},
			},
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
								"storage": pulumi.String("256Mi"),
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

func service(ctx *pulumi.Context) error {
	_, err := corev1.NewService(ctx, "tigerbeetle-service", &corev1.ServiceArgs{
		ApiVersion: pulumi.String("v1"),
		Kind:       pulumi.String("Service"),
		Metadata: &metav1.ObjectMetaArgs{
			Name: pulumi.String("tigerbeetle"),
			Labels: pulumi.StringMap{
				"app": pulumi.String("tigerbeetle"),
			},
		},
		Spec: &corev1.ServiceSpecArgs{
			Ports: corev1.ServicePortArray{
				&corev1.ServicePortArgs{
					Port:       pulumi.Int(80),
					TargetPort: pulumi.Int(8080),
					Name:       pulumi.String("http"),
				},
			},
			ClusterIP: pulumi.String("None"),
			Selector: pulumi.StringMap{
				"app": pulumi.String("tigerbeetle"),
			},
		},
	})
	if err != nil {
		return err
	}
	return nil
}

func rbac(ctx *pulumi.Context) error {
	_, err := corev1.NewServiceAccount(ctx, "tigerbeetle-sa", &corev1.ServiceAccountArgs{
		ApiVersion: pulumi.String("v1"),
		Kind:       pulumi.String("ServiceAccount"),
		Metadata: &metav1.ObjectMetaArgs{
			Name: pulumi.String("tigerbeetle"),
			Labels: pulumi.StringMap{
				"app": pulumi.String("tigerbeetle"),
			},
		},
		AutomountServiceAccountToken: pulumi.Bool(false),
	})
	if err != nil {
		return err
	}

	_, err = rbacv1.NewRole(ctx, "tigerbeetle-role", &rbacv1.RoleArgs{
		ApiVersion: pulumi.String("rbac.authorization.k8s.io/v1"),
		Kind:       pulumi.String("Role"),
		Metadata: &metav1.ObjectMetaArgs{
			Name: pulumi.String("tigerbeetle"),
			Labels: pulumi.StringMap{
				"app": pulumi.String("tigerbeetle"),
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

	_, err = rbacv1.NewRoleBinding(ctx, "tigerbeetle-rb", &rbacv1.RoleBindingArgs{
		ApiVersion: pulumi.String("rbac.authorization.k8s.io/v1"),
		Kind:       pulumi.String("RoleBinding"),
		Metadata: &metav1.ObjectMetaArgs{
			Name: pulumi.String("tigerbeetle"),
			Labels: pulumi.StringMap{
				"app": pulumi.String("tigerbeetle"),
			},
		},
		RoleRef: &rbacv1.RoleRefArgs{
			ApiGroup: pulumi.String("rbac.authorization.k8s.io"),
			Kind:     pulumi.String("Role"),
			Name:     pulumi.String("tigerbeetle"),
		},
		Subjects: rbacv1.SubjectArray{
			&rbacv1.SubjectArgs{
				Kind:      pulumi.String("ServiceAccount"),
				Name:      pulumi.String("tigerbeetle"),
				Namespace: pulumi.String("default"),
			},
		},
	})
	if err != nil {
		return err
	}
	return nil
}
