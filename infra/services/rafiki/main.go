package rafiki

import (
	appsv1 "github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes/apps/v1"
	corev1 "github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes/core/v1"
	metav1 "github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes/meta/v1"
	rbacv1 "github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes/rbac/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"gitlab.com/fynbos/infra/services/ingress"
)

type DeployRafikiArgs struct {
	DbUrl                      pulumi.StringInput
	TbClusterID                string
	TbReplicaAddresses         string
	IlpAddress                 string
	NonceRedisKey              string
	RedisUrl                   string
	AuthServerGrantUrl         string
	AuthServerIntrospectionUrl string
	StreamSecret               pulumi.StringInput
	AdminKey                   pulumi.StringInput
	Hostname                   string
	PublicHost                 string
	WebhookUrl                 string
	ImageRepo                  string
	ImageTag                   string
}

func DeployRafiki(ctx *pulumi.Context, args *DeployRafikiArgs) error {
	if err := deployService(ctx); err != nil {
		return err
	}

	if err := deployRbac(ctx); err != nil {
		return err
	}

	if err := deployDeployment(ctx, args); err != nil {
		return err
	}

	if err := deployIngress(ctx, args.Hostname); err != nil {
		return err
	}

	return nil
}

func deployDeployment(ctx *pulumi.Context, args *DeployRafikiArgs) error {
	// store admin key and stream secret in a k8s secret.
	streamSecret, err := corev1.NewSecret(ctx, "rafiki-stream-secret", &corev1.SecretArgs{
		ApiVersion: pulumi.String("v1"),
		Kind:       pulumi.String("Secret"),
		Metadata: &metav1.ObjectMetaArgs{
			Name: pulumi.String("rafiki-stream-secret"),
			Labels: pulumi.StringMap{
				"app": pulumi.String("rafiki"),
			},
		},
		StringData: pulumi.StringMap{
			"streamSecret": args.StreamSecret,
		},
	})
	if err != nil {
		return err
	}

	adminKey, err := corev1.NewSecret(ctx, "rafiki-admin-key", &corev1.SecretArgs{
		ApiVersion: pulumi.String("v1"),
		Kind:       pulumi.String("Secret"),
		Metadata: &metav1.ObjectMetaArgs{
			Name: pulumi.String("rafiki-admin-key"),
			Labels: pulumi.StringMap{
				"app": pulumi.String("rafiki"),
			},
		},
		StringData: pulumi.StringMap{
			"adminKey": args.AdminKey,
		},
	})
	if err != nil {
		return err
	}

	_, err = appsv1.NewDeployment(ctx, "rafiki", &appsv1.DeploymentArgs{
		ApiVersion: pulumi.String("apps/v1"),
		Kind:       pulumi.String("Deployment"),
		Metadata: &metav1.ObjectMetaArgs{
			Name: pulumi.String("rafiki"),
			Labels: pulumi.StringMap{
				"app": pulumi.String("rafiki"),
			},
		},
		Spec: &appsv1.DeploymentSpecArgs{
			Replicas: pulumi.Int(1),
			Selector: &metav1.LabelSelectorArgs{
				MatchLabels: pulumi.StringMap{
					"app": pulumi.String("rafiki"),
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
						"app": pulumi.String("rafiki"),
					},
				},
				Spec: &corev1.PodSpecArgs{
					SecurityContext: &corev1.PodSecurityContextArgs{
						RunAsUser: pulumi.Int(65532),
					},
					// Volumes: corev1.VolumeArray{
					// 	&corev1.VolumeArgs{
					// 		Name: pulumi.String("cockroach-certs"),
					// 		Secret: &corev1.SecretVolumeSourceArgs{
					// 			SecretName: pulumi.String("cockroachdb-backend"),
					// 			Items: corev1.KeyToPathArray{
					// 				&corev1.KeyToPathArgs{
					// 					Key:  pulumi.String("tls.key"),
					// 					Path: pulumi.String("client.backend.key"),
					// 				},
					// 				&corev1.KeyToPathArgs{
					// 					Key:  pulumi.String("tls.crt"),
					// 					Path: pulumi.String("client.backend.crt"),
					// 				},
					// 				&corev1.KeyToPathArgs{
					// 					Key:  pulumi.String("ca.crt"),
					// 					Path: pulumi.String("ca.crt"),
					// 				},
					// 			},
					// 		},
					// 	},
					// },
					ServiceAccountName: pulumi.String("rafiki"),
					Containers: corev1.ContainerArray{
						&corev1.ContainerArgs{
							Name:            pulumi.String("rafiki"),
							Image:           pulumi.Sprintf("%s:%s", args.ImageRepo, args.ImageTag),
							ImagePullPolicy: pulumi.String("Always"),
							Ports: corev1.ContainerPortArray{
								&corev1.ContainerPortArgs{
									ContainerPort: pulumi.Int(3001),
									Name:          pulumi.String("http"),
								},
								&corev1.ContainerPortArgs{
									ContainerPort: pulumi.Int(3002),
									Name:          pulumi.String("ilp"),
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
								PeriodSeconds: pulumi.Int(5),
							},
							Env: corev1.EnvVarArray{
								&corev1.EnvVarArgs{
									Name:  pulumi.String("DATABASE_URL"),
									Value: args.DbUrl,
								},
								&corev1.EnvVarArgs{
									Name:  pulumi.String("TIGERBEETLE_CLUSTER_ID"),
									Value: pulumi.String(args.TbClusterID),
								},
								&corev1.EnvVarArgs{
									Name:  pulumi.String("TIGERBEETLE_REPLICA_ADDRESSES"),
									Value: pulumi.String(args.TbReplicaAddresses),
								},
								&corev1.EnvVarArgs{
									Name:  pulumi.String("REDIS_URL"),
									Value: pulumi.String(args.RedisUrl),
								},
								&corev1.EnvVarArgs{
									Name:  pulumi.String("NONCE_REDIS_KEY"),
									Value: pulumi.String(args.NonceRedisKey),
								},
								&corev1.EnvVarArgs{
									Name:  pulumi.String("AUTH_SERVER_GRANT_URL"),
									Value: pulumi.String(args.AuthServerGrantUrl),
								},
								&corev1.EnvVarArgs{
									Name:  pulumi.String("AUTH_SERVER_INTROSPECTION_URL"),
									Value: pulumi.String(args.AuthServerIntrospectionUrl),
								},
								&corev1.EnvVarArgs{
									Name:  pulumi.String("ILP_ADDRESS"),
									Value: pulumi.String(args.IlpAddress),
								},
								&corev1.EnvVarArgs{
									Name: pulumi.String("STREAM_SECRET"),
									ValueFrom: corev1.EnvVarSourceArgs{
										SecretKeyRef: corev1.SecretKeySelectorArgs{
											Name: pulumi.String("rafiki-stream-secret"),
											Key:  pulumi.String("streamSecret"),
										},
									},
								},
								&corev1.EnvVarArgs{
									Name: pulumi.String("ADMIN_KEY"),
									ValueFrom: corev1.EnvVarSourceArgs{
										SecretKeyRef: corev1.SecretKeySelectorArgs{
											Name: pulumi.String("rafiki-admin-key"),
											Key:  pulumi.String("adminKey"),
										},
									},
								},
								&corev1.EnvVarArgs{
									Name:  pulumi.String("PUBLIC_HOST"),
									Value: pulumi.String(args.PublicHost),
								},
								&corev1.EnvVarArgs{
									Name:  pulumi.String("LOG_LEVEL"),
									Value: pulumi.String("debug"),
								},
								&corev1.EnvVarArgs{
									Name:  pulumi.String("WEBHOOK_URL"),
									Value: pulumi.String(args.WebhookUrl),
								},
							},
						},
					},
					TerminationGracePeriodSeconds: pulumi.Int(30),
				},
			},
		},
	}, pulumi.DependsOn([]pulumi.Resource{streamSecret, adminKey}))
	if err != nil {
		return err
	}

	return nil
}

func deployRbac(ctx *pulumi.Context) error {
	_, err := corev1.NewServiceAccount(ctx, "rafiki-sa", &corev1.ServiceAccountArgs{
		ApiVersion: pulumi.String("v1"),
		Kind:       pulumi.String("ServiceAccount"),
		Metadata: &metav1.ObjectMetaArgs{
			Name: pulumi.String("rafiki"),
			Labels: pulumi.StringMap{
				"app": pulumi.String("rafiki"),
			},
		},
		AutomountServiceAccountToken: pulumi.Bool(false),
	})
	if err != nil {
		return err
	}
	_, err = rbacv1.NewRole(ctx, "rafiki-role", &rbacv1.RoleArgs{
		ApiVersion: pulumi.String("rbac.authorization.k8s.io/v1"),
		Kind:       pulumi.String("Role"),
		Metadata: &metav1.ObjectMetaArgs{
			Name: pulumi.String("rafiki"),
			Labels: pulumi.StringMap{
				"app": pulumi.String("rafiki"),
			},
		},
	})
	if err != nil {
		return err
	}
	_, err = rbacv1.NewRoleBinding(ctx, "rafiki-rb", &rbacv1.RoleBindingArgs{
		ApiVersion: pulumi.String("rbac.authorization.k8s.io/v1"),
		Kind:       pulumi.String("RoleBinding"),
		Metadata: &metav1.ObjectMetaArgs{
			Name: pulumi.String("rafiki"),
			Labels: pulumi.StringMap{
				"app": pulumi.String("rafiki"),
			},
		},
		RoleRef: &rbacv1.RoleRefArgs{
			ApiGroup: pulumi.String("rbac.authorization.k8s.io"),
			Kind:     pulumi.String("Role"),
			Name:     pulumi.String("rafiki"),
		},
		Subjects: rbacv1.SubjectArray{
			&rbacv1.SubjectArgs{
				Kind:      pulumi.String("ServiceAccount"),
				Name:      pulumi.String("rafiki"),
				Namespace: pulumi.String("default"),
			},
		},
	})
	if err != nil {
		return err
	}
	return nil
}

func deployService(ctx *pulumi.Context) error {
	_, err := corev1.NewService(ctx, "rafiki-service", &corev1.ServiceArgs{
		ApiVersion: pulumi.String("v1"),
		Kind:       pulumi.String("Service"),
		Metadata: &metav1.ObjectMetaArgs{
			Name: pulumi.String("rafiki"),
			Labels: pulumi.StringMap{
				"app": pulumi.String("rafiki"),
			},
		},
		Spec: &corev1.ServiceSpecArgs{
			Ports: corev1.ServicePortArray{
				&corev1.ServicePortArgs{
					Port:       pulumi.Int(80),
					TargetPort: pulumi.Int(3001),
					Name:       pulumi.String("http"),
				},
			},
			Selector: pulumi.StringMap{
				"app": pulumi.String("rafiki"),
			},
		},
	})
	if err != nil {
		return err
	}

	_, err = corev1.NewService(ctx, "connector-service", &corev1.ServiceArgs{
		ApiVersion: pulumi.String("v1"),
		Kind:       pulumi.String("Service"),
		Metadata: &metav1.ObjectMetaArgs{
			Name: pulumi.String("connector"),
			Labels: pulumi.StringMap{
				"app": pulumi.String("connector"),
			},
		},
		Spec: &corev1.ServiceSpecArgs{
			Ports: corev1.ServicePortArray{
				&corev1.ServicePortArgs{
					Port:       pulumi.Int(80),
					TargetPort: pulumi.Int(3002),
					Name:       pulumi.String("ilp"),
				},
			},
			Selector: pulumi.StringMap{
				"app": pulumi.String("rafiki"),
			},
		},
	})
	if err != nil {
		return err
	}

	return nil
}

func deployIngress(ctx *pulumi.Context, hn string) error {
	err := ingress.DeployMapping(ctx, &ingress.MappingArgs{
		Name:     "rafiki",
		Hostname: hn,
		Prefix:   "/",
		Rewrite:  "/",
		Service:  "rafiki",
	})
	if err != nil {
		return err
	}

	err = ingress.DeployMapping(ctx, &ingress.MappingArgs{
		Name:     "rafiki-graphql",
		Hostname: hn,
		Prefix:   "/graphql",
		Rewrite:  "/graphql",
		Service:  "rafiki",
	})
	if err != nil {
		return err
	}

	return nil
}
