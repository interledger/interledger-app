package backend

import (
	"github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes/apiextensions"
	appsv1 "github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes/apps/v1"
	corev1 "github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes/core/v1"
	metav1 "github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes/meta/v1"
	rbacv1 "github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes/rbac/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"gitlab.com/fynbos/infra/services/ingress"
)

type DeployBackendArgs struct {
	Cert      *apiextensions.CustomResource
	ImageRepo string
	ImageTag  string
}

func DeployBackend(ctx *pulumi.Context, args DeployBackendArgs) error {

	err := deployService(ctx)
	if err != nil {
		return err
	}
	err = deployIngress(ctx)
	if err != nil {
		return err
	}
	err = deployRbac(ctx)
	if err != nil {
		return err
	}
	err = deployDeployment(ctx, args.ImageRepo, args.ImageTag, args.Cert)
	if err != nil {
		return err
	}

	return nil
}

func deployService(ctx *pulumi.Context) error {
	_, err := corev1.NewService(ctx, "backend-service", &corev1.ServiceArgs{
		ApiVersion: pulumi.String("v1"),
		Kind:       pulumi.String("Service"),
		Metadata: &metav1.ObjectMetaArgs{
			Name: pulumi.String("backend"),
			Labels: pulumi.StringMap{
				"app": pulumi.String("backend"),
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
			Selector: pulumi.StringMap{
				"app": pulumi.String("backend"),
			},
		},
	})
	if err != nil {
		return err
	}
	return nil
}

func deployDeployment(ctx *pulumi.Context, imageRepo string, imageTag string, cert *apiextensions.CustomResource) error {
	_, err := appsv1.NewDeployment(ctx, "backend-deployment", &appsv1.DeploymentArgs{
		ApiVersion: pulumi.String("apps/v1"),
		Kind:       pulumi.String("Deployment"),
		Metadata: &metav1.ObjectMetaArgs{
			Name: pulumi.String("backend"),
			Labels: pulumi.StringMap{
				"app": pulumi.String("backend"),
			},
		},
		Spec: &appsv1.DeploymentSpecArgs{
			Replicas: pulumi.Int(1),
			Selector: &metav1.LabelSelectorArgs{
				MatchLabels: pulumi.StringMap{
					"app": pulumi.String("backend"),
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
						"app": pulumi.String("backend"),
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
								SecretName: pulumi.String("cockroachdb-backend"),
								Items: corev1.KeyToPathArray{
									&corev1.KeyToPathArgs{
										Key:  pulumi.String("tls.key"),
										Path: pulumi.String("client.backend.key"),
									},
									&corev1.KeyToPathArgs{
										Key:  pulumi.String("tls.crt"),
										Path: pulumi.String("client.backend.crt"),
									},
									&corev1.KeyToPathArgs{
										Key:  pulumi.String("ca.crt"),
										Path: pulumi.String("ca.crt"),
									},
								},
							},
						},
					},
					ServiceAccountName: pulumi.String("backend"),
					//InitContainers: corev1.ContainerArray{
					//	&corev1.ContainerArgs{
					//		Name:            pulumi.String("init-db"),
					//		Image: pulumi.Sprintf("%s/backend", imageRepo),
					//		ImagePullPolicy: pulumi.String("Always"),
					//		Env: corev1.EnvVarArray{
					//			&corev1.EnvVarArgs{
					//				Name:  pulumi.String("DB_URL"),
					//				Value: pulumi.String("cockroach://backend@cockroachdb-public:26257/backend?sslmode=verify-full&max_conns=20&max_idle_conns=4"),
					//			},
					//		},
					//		VolumeMounts: corev1.VolumeMountArray{
					//			&corev1.VolumeMountArgs{
					//				Name:      pulumi.String("cockroach-certs"),
					//				MountPath: pulumi.String("/cockroach-certs"),
					//			},
					//		},
					//	},
					//},
					Containers: corev1.ContainerArray{
						&corev1.ContainerArgs{
							Name:            pulumi.String("backend"),
							Image:           pulumi.Sprintf("%s/backend:%s", imageRepo, imageTag),
							ImagePullPolicy: pulumi.String("Always"),
							Ports: corev1.ContainerPortArray{
								&corev1.ContainerPortArgs{
									ContainerPort: pulumi.Int(8080),
									Name:          pulumi.String("http"),
								},
							},
							LivenessProbe: &corev1.ProbeArgs{
								HttpGet: &corev1.HTTPGetActionArgs{
									Path:   pulumi.String("/healthz"),
									Port:   pulumi.String("http"),
									Scheme: pulumi.String("HTTP"),
								},
								PeriodSeconds: pulumi.Int(5),
							},
							ReadinessProbe: &corev1.ProbeArgs{
								HttpGet: &corev1.HTTPGetActionArgs{
									Path:   pulumi.String("/healthz"),
									Port:   pulumi.String("http"),
									Scheme: pulumi.String("HTTP"),
								},
								PeriodSeconds:    pulumi.Int(5),
								FailureThreshold: pulumi.Int(2),
							},
							Env: corev1.EnvVarArray{
								&corev1.EnvVarArgs{
									Name:  pulumi.String("DB_URL"),
									Value: pulumi.String("cockroach://backend@cockroachdb-public:26257/backend?sslmode=verify-full&max_conns=20&max_idle_conns=4"),
								},
								&corev1.EnvVarArgs{
									Name:  pulumi.String("KRATOS_URL"),
									Value: pulumi.String("http://kratos-public"),
								},
								&corev1.EnvVarArgs{
									Name:  pulumi.String("LOG_LEVEL"),
									Value: pulumi.String("info"),
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

func deployRbac(ctx *pulumi.Context) error {
	_, err := corev1.NewServiceAccount(ctx, "backend-sa", &corev1.ServiceAccountArgs{
		ApiVersion: pulumi.String("v1"),
		Kind:       pulumi.String("ServiceAccount"),
		Metadata: &metav1.ObjectMetaArgs{
			Name: pulumi.String("backend"),
			Labels: pulumi.StringMap{
				"app": pulumi.String("backend"),
			},
		},
		AutomountServiceAccountToken: pulumi.Bool(false),
	})
	if err != nil {
		return err
	}
	_, err = rbacv1.NewRole(ctx, "backend-role", &rbacv1.RoleArgs{
		ApiVersion: pulumi.String("rbac.authorization.k8s.io/v1"),
		Kind:       pulumi.String("Role"),
		Metadata: &metav1.ObjectMetaArgs{
			Name: pulumi.String("backend"),
			Labels: pulumi.StringMap{
				"app": pulumi.String("backend"),
			},
		},
	})
	if err != nil {
		return err
	}
	_, err = rbacv1.NewRoleBinding(ctx, "backend-rb", &rbacv1.RoleBindingArgs{
		ApiVersion: pulumi.String("rbac.authorization.k8s.io/v1"),
		Kind:       pulumi.String("RoleBinding"),
		Metadata: &metav1.ObjectMetaArgs{
			Name: pulumi.String("backend"),
			Labels: pulumi.StringMap{
				"app": pulumi.String("backend"),
			},
		},
		RoleRef: &rbacv1.RoleRefArgs{
			ApiGroup: pulumi.String("rbac.authorization.k8s.io"),
			Kind:     pulumi.String("Role"),
			Name:     pulumi.String("backend"),
		},
		Subjects: rbacv1.SubjectArray{
			&rbacv1.SubjectArgs{
				Kind:      pulumi.String("ServiceAccount"),
				Name:      pulumi.String("backend"),
				Namespace: pulumi.String("default"),
			},
		},
	})
	if err != nil {
		return err
	}
	return nil
}

func deployIngress(ctx *pulumi.Context) error {
	err := ingress.DeployMapping(ctx, &ingress.MappingArgs{
		Name:     "backend-graphql",
		Hostname: "*",
		Prefix:   "/graphql",
		Service:  "backend",
	})
	if err != nil {
		return err
	}

	return nil
}
