package pacioli

import (
	"github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes/apiextensions"
	appsv1 "github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes/apps/v1"
	corev1 "github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes/core/v1"
	metav1 "github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes/meta/v1"
	policyv1 "github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes/policy/v1"
	rbacv1 "github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes/rbac/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

const name = "pacioli"

type DeployPacioliArgs struct {
	Cert      *apiextensions.CustomResource
	ImageRepo string
	ImageTag  string
	Namespace string
}

func DeployPacioli(ctx *pulumi.Context, args *DeployPacioliArgs) error {
	err := deployService(ctx)
	if err != nil {
		return err
	}

	err = deployRbac(ctx, args.Namespace)
	if err != nil {
		return err
	}

	err = deployPodDisruptionBudget(ctx)
	if err != nil {
		return err
	}

	err = deployDeployment(ctx, args.ImageRepo, args.ImageTag, args.Cert)
	if err != nil {
		return err
	}

	return nil
}

func deployPodDisruptionBudget(ctx *pulumi.Context) error {
	_, err := policyv1.NewPodDisruptionBudget(ctx, name+"-pdb", &policyv1.PodDisruptionBudgetArgs{
		ApiVersion: pulumi.String("policy/v1"),
		Kind:       pulumi.String("PodDisruptionBudget"),
		Metadata: &metav1.ObjectMetaArgs{
			Name: pulumi.String(name),
			Labels: pulumi.StringMap{
				"app": pulumi.String(name),
			},
		},
		Spec: &policyv1.PodDisruptionBudgetSpecArgs{
			Selector: metav1.LabelSelectorArgs{
				MatchLabels: pulumi.StringMap{
					"app": pulumi.String(name),
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

func deployService(ctx *pulumi.Context) error {
	_, err := corev1.NewService(ctx, name+"-service", &corev1.ServiceArgs{
		ApiVersion: pulumi.String("v1"),
		Kind:       pulumi.String("Service"),
		Metadata: &metav1.ObjectMetaArgs{
			Name: pulumi.String(name),
			Labels: pulumi.StringMap{
				"app": pulumi.String(name),
			},
		},
		Spec: &corev1.ServiceSpecArgs{
			Ports: corev1.ServicePortArray{
				&corev1.ServicePortArgs{
					Port:       pulumi.Int(443), // default grpc port.
					TargetPort: pulumi.Int(8080),
					Name:       pulumi.String("http"),
				},
			},
			Selector: pulumi.StringMap{
				"app": pulumi.String(name),
			},
		},
	})
	if err != nil {
		return err
	}
	return nil
}

func deployRbac(ctx *pulumi.Context, namespace string) error {
	_, err := corev1.NewServiceAccount(ctx, name+"-sa", &corev1.ServiceAccountArgs{
		ApiVersion: pulumi.String("v1"),
		Kind:       pulumi.String("ServiceAccount"),
		Metadata: &metav1.ObjectMetaArgs{
			Name: pulumi.String(name),
			Labels: pulumi.StringMap{
				"app": pulumi.String(name),
			},
		},
		AutomountServiceAccountToken: pulumi.Bool(false),
	})
	if err != nil {
		return err
	}
	_, err = rbacv1.NewRole(ctx, name+"-role", &rbacv1.RoleArgs{
		ApiVersion: pulumi.String("rbac.authorization.k8s.io/v1"),
		Kind:       pulumi.String("Role"),
		Metadata: &metav1.ObjectMetaArgs{
			Name: pulumi.String(name),
			Labels: pulumi.StringMap{
				"app": pulumi.String(name),
			},
		},
	})
	if err != nil {
		return err
	}
	_, err = rbacv1.NewRoleBinding(ctx, name+"-rb", &rbacv1.RoleBindingArgs{
		ApiVersion: pulumi.String("rbac.authorization.k8s.io/v1"),
		Kind:       pulumi.String("RoleBinding"),
		Metadata: &metav1.ObjectMetaArgs{
			Name: pulumi.String(name),
			Labels: pulumi.StringMap{
				"app": pulumi.String(name),
			},
		},
		RoleRef: &rbacv1.RoleRefArgs{
			ApiGroup: pulumi.String("rbac.authorization.k8s.io"),
			Kind:     pulumi.String("Role"),
			Name:     pulumi.String(name),
		},
		Subjects: rbacv1.SubjectArray{
			&rbacv1.SubjectArgs{
				Kind:      pulumi.String("ServiceAccount"),
				Name:      pulumi.String(name),
				Namespace: pulumi.String(namespace),
			},
		},
	})
	if err != nil {
		return err
	}
	return nil
}

func deployDeployment(ctx *pulumi.Context, imageRepo string, imageTag string, cert *apiextensions.CustomResource) error {
	_, err := appsv1.NewDeployment(ctx, name+"-deployment", &appsv1.DeploymentArgs{
		ApiVersion: pulumi.String("apps/v1"),
		Kind:       pulumi.String("Deployment"),
		Metadata: &metav1.ObjectMetaArgs{
			Name: pulumi.String(name),
			Labels: pulumi.StringMap{
				"app": pulumi.String(name),
			},
		},
		Spec: &appsv1.DeploymentSpecArgs{
			Replicas: pulumi.Int(1),
			Selector: &metav1.LabelSelectorArgs{
				MatchLabels: pulumi.StringMap{
					"app": pulumi.String(name),
				},
			},
			Strategy: &appsv1.DeploymentStrategyArgs{
				Type: pulumi.String("RollingUpdate"),
				RollingUpdate: &appsv1.RollingUpdateDeploymentArgs{
					MaxSurge:       pulumi.Int(2),
					MaxUnavailable: pulumi.Int(1),
				},
			},
			Template: &corev1.PodTemplateSpecArgs{
				Metadata: &metav1.ObjectMetaArgs{
					Labels: pulumi.StringMap{
						"app": pulumi.String(name),
					},
				},
				Spec: &corev1.PodSpecArgs{
					SecurityContext: &corev1.PodSecurityContextArgs{
						RunAsUser: pulumi.Int(65532),
					},
					Volumes: corev1.VolumeArray{
						&corev1.VolumeArgs{
							Name: pulumi.String("cockroach-certs"),
							Secret: &corev1.SecretVolumeSourceArgs{
								SecretName: pulumi.String("cockroachdb-" + name),
								Items: corev1.KeyToPathArray{
									&corev1.KeyToPathArgs{
										Key:  pulumi.String("tls.key"),
										Path: pulumi.String("client." + name + ".key"),
									},
									&corev1.KeyToPathArgs{
										Key:  pulumi.String("tls.crt"),
										Path: pulumi.String("client." + name + ".crt"),
									},
									&corev1.KeyToPathArgs{
										Key:  pulumi.String("ca.crt"),
										Path: pulumi.String("ca.crt"),
									},
								},
							},
						},
					},
					ServiceAccountName: pulumi.String(name),
					InitContainers: corev1.ContainerArray{
						&corev1.ContainerArgs{
							Name:            pulumi.String(name + "-init"),
							Image:           pulumi.Sprintf("%s/%s:%s", imageRepo, name, imageTag),
							ImagePullPolicy: pulumi.String("Always"),
							Env: corev1.EnvVarArray{
								&corev1.EnvVarArgs{
									Name:  pulumi.String("DB_URL"),
									Value: pulumi.Sprintf("cockroach://%s@cockroachdb-public:26257/%s?sslmode=verify-full&max_conns=20&max_idle_conns=4", name, name),
								},
								&corev1.EnvVarArgs{
									Name:  pulumi.String("TB_URL"),
									Value: pulumi.String("tigerbeetle-0.tigerbeetle.default.svc.cluster.local"),
								},
								&corev1.EnvVarArgs{
									Name:  pulumi.String("TB_CLUSTER_ID"),
									Value: pulumi.String("0"),
								},
							},
							Args: pulumi.StringArray{pulumi.String("init")},
							VolumeMounts: corev1.VolumeMountArray{
								&corev1.VolumeMountArgs{
									Name:      pulumi.String("cockroach-certs"),
									MountPath: pulumi.String("/cockroach-certs"),
								},
							},
						},
					},
					Containers: corev1.ContainerArray{
						&corev1.ContainerArgs{
							Name:            pulumi.String(name),
							Image:           pulumi.Sprintf("%s/%s:%s", imageRepo, name, imageTag),
							ImagePullPolicy: pulumi.String("Always"),
							Args:            pulumi.StringArray{pulumi.String("start")},
							Ports: corev1.ContainerPortArray{
								&corev1.ContainerPortArgs{
									ContainerPort: pulumi.Int(8080),
									Name:          pulumi.String("http"),
								},
							},
							LivenessProbe: &corev1.ProbeArgs{
								Exec: &corev1.ExecActionArgs{
									Command: pulumi.StringArray{
										pulumi.String("/dist/grpc_health_probe"),
										pulumi.String("-addr=:8080"),
										pulumi.String("-service=pacioli"),
									},
								},
								PeriodSeconds: pulumi.Int(5),
							},
							ReadinessProbe: &corev1.ProbeArgs{
								Exec: &corev1.ExecActionArgs{
									Command: pulumi.StringArray{
										pulumi.String("/dist/grpc_health_probe"),
										pulumi.String("-addr=:8080"),
										pulumi.String("-service=pacioli"),
									},
								},
								PeriodSeconds:    pulumi.Int(5),
								FailureThreshold: pulumi.Int(2),
							},
							Env: corev1.EnvVarArray{
								&corev1.EnvVarArgs{
									Name:  pulumi.String("PORT"),
									Value: pulumi.String("8080"),
								},
								&corev1.EnvVarArgs{
									Name:  pulumi.String("DB_URL"),
									Value: pulumi.Sprintf("cockroach://%s@cockroachdb-public:26257/%s?sslmode=verify-full&max_conns=20&max_idle_conns=4", name, name),
								},
								&corev1.EnvVarArgs{
									Name:  pulumi.String("TB_URL"),
									Value: pulumi.String("tigerbeetle-0.tigerbeetle.default.svc.cluster.local"),
								},
								&corev1.EnvVarArgs{
									Name:  pulumi.String("TB_CLUSTER_ID"),
									Value: pulumi.String("0"),
								},
							},
							VolumeMounts: corev1.VolumeMountArray{
								&corev1.VolumeMountArgs{
									Name:      pulumi.String("cockroach-certs"),
									MountPath: pulumi.String("/cockroach-certs"),
								},
							},
						},
					},
					TerminationGracePeriodSeconds: pulumi.Int(30),
				},
			},
		},
	}, pulumi.DependsOn([]pulumi.Resource{cert}))
	if err != nil {
		return err
	}
	return nil
}
