package rafiki

import (
	"fmt"

	"github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes/apiextensions"
	appsv1 "github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes/apps/v1"
	corev1 "github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes/core/v1"
	metav1 "github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes/meta/v1"
	rbacv1 "github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes/rbac/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"gitlab.com/fynbos/infra/services/ingress"
)

type DeployRafikiArgs struct {
	Name                       string
	DeployPlaygroundIngress    bool
	DbCert                     *apiextensions.CustomResource
	DbBaseUrl                  string
	TbClusterID                string
	TbReplicaAddresses         string
	IlpAddress                 string
	NonceRedisKey              string
	RedisUrl                   string
	RedisCert                  *apiextensions.CustomResource
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
	if err := deployService(ctx, args.Name); err != nil {
		return err
	}

	if err := deployRbac(ctx, args.Name); err != nil {
		return err
	}

	if err := deployDeployment(ctx, args); err != nil {
		return err
	}

	if err := deployIngress(ctx, args.Hostname, args.Name); err != nil {
		return err
	}

	if args.DeployPlaygroundIngress {
		if err := deployPlaygroundIngress(ctx, args.Hostname, args.Name); err != nil {
			return err
		}
	}

	return nil
}

func deployDeployment(ctx *pulumi.Context, args *DeployRafikiArgs) error {
	// store admin key and stream secret in a k8s secret.
	streamSecret, err := corev1.NewSecret(ctx, fmt.Sprintf("%s-stream-secret", args.Name), &corev1.SecretArgs{
		ApiVersion: pulumi.String("v1"),
		Kind:       pulumi.String("Secret"),
		Metadata: &metav1.ObjectMetaArgs{
			Name: pulumi.String(fmt.Sprintf("%s-stream-secret", args.Name)),
			Labels: pulumi.StringMap{
				"app": pulumi.String(args.Name),
			},
		},
		StringData: pulumi.StringMap{
			"streamSecret": args.StreamSecret,
		},
	})
	if err != nil {
		return err
	}

	adminKey, err := corev1.NewSecret(ctx, fmt.Sprintf("%s-admin-key", args.Name), &corev1.SecretArgs{
		ApiVersion: pulumi.String("v1"),
		Kind:       pulumi.String("Secret"),
		Metadata: &metav1.ObjectMetaArgs{
			Name: pulumi.String(fmt.Sprintf("%s-admin-key", args.Name)),
			Labels: pulumi.StringMap{
				"app": pulumi.String(args.Name),
			},
		},
		StringData: pulumi.StringMap{
			"adminKey": args.AdminKey,
		},
	})
	if err != nil {
		return err
	}

	_, err = appsv1.NewDeployment(ctx, args.Name, &appsv1.DeploymentArgs{
		ApiVersion: pulumi.String("apps/v1"),
		Kind:       pulumi.String("Deployment"),
		Metadata: &metav1.ObjectMetaArgs{
			Name: pulumi.String(args.Name),
			Labels: pulumi.StringMap{
				"app": pulumi.String(args.Name),
			},
		},
		Spec: &appsv1.DeploymentSpecArgs{
			Replicas: pulumi.Int(1),
			Selector: &metav1.LabelSelectorArgs{
				MatchLabels: pulumi.StringMap{
					"app": pulumi.String(args.Name),
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
						"app": pulumi.String(args.Name),
					},
				},
				Spec: &corev1.PodSpecArgs{
					SecurityContext: &corev1.PodSecurityContextArgs{
						RunAsUser: pulumi.Int(65532),
					},
					Volumes: corev1.VolumeArray{
						&corev1.VolumeArgs{
							Name: pulumi.String("postgres-certs"),
							Secret: &corev1.SecretVolumeSourceArgs{
								SecretName: pulumi.String(fmt.Sprintf("postgresdb-%s", args.Name)),
								Items: corev1.KeyToPathArray{
									&corev1.KeyToPathArgs{
										Key:  pulumi.String("tls.key"),
										Path: pulumi.String("tls.key"),
									},
									&corev1.KeyToPathArgs{
										Key:  pulumi.String("tls.crt"),
										Path: pulumi.String("tls.crt"),
									},
									&corev1.KeyToPathArgs{
										Key:  pulumi.String("ca.crt"),
										Path: pulumi.String("ca.crt"),
									},
								},
							},
						},
						&corev1.VolumeArgs{
							Name: pulumi.String("redis-certs"),
							Secret: &corev1.SecretVolumeSourceArgs{
								SecretName: pulumi.String(fmt.Sprintf("redis-%s", args.Name)),
								Items: corev1.KeyToPathArray{
									&corev1.KeyToPathArgs{
										Key:  pulumi.String("tls.key"),
										Path: pulumi.String("tls.key"),
									},
									&corev1.KeyToPathArgs{
										Key:  pulumi.String("tls.crt"),
										Path: pulumi.String("tls.crt"),
									},
									&corev1.KeyToPathArgs{
										Key:  pulumi.String("ca.crt"),
										Path: pulumi.String("ca.crt"),
									},
								},
							},
						},
					},
					ServiceAccountName: pulumi.String(args.Name),
					Containers: corev1.ContainerArray{
						&corev1.ContainerArgs{
							Name:            pulumi.String(args.Name),
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
									Name:  pulumi.String("PORT"),
									Value: pulumi.String("3001"),
								},
								&corev1.EnvVarArgs{
									Name:  pulumi.String("CONNECTOR_PORT"),
									Value: pulumi.String("3002"),
								},
								&corev1.EnvVarArgs{
									Name: pulumi.String("DATABASE_URL"),
									Value: pulumi.String(
										args.DbBaseUrl +
											"?sslmode=verify-full&" +
											"sslcert=/postgres-certs/tls.crt&" +
											"sslkey=/postgres-certs/tls.key&" +
											"sslrootcert=/postgres-certs/ca.crt",
									),
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
									Name:  pulumi.String("REDIS_TLS_CA_FILE_PATH"),
									Value: pulumi.String("/redis-certs/ca.crt"),
								},
								&corev1.EnvVarArgs{
									Name:  pulumi.String("REDIS_TLS_KEY_FILE_PATH"),
									Value: pulumi.String("/redis-certs/tls.key"),
								},
								&corev1.EnvVarArgs{
									Name:  pulumi.String("REDIS_TLS_CERT_FILE_PATH"),
									Value: pulumi.String("/redis-certs/tls.crt"),
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
											Name: pulumi.String(fmt.Sprintf("%s-stream-secret", args.Name)),
											Key:  pulumi.String("streamSecret"),
										},
									},
								},
								&corev1.EnvVarArgs{
									Name: pulumi.String("ADMIN_KEY"),
									ValueFrom: corev1.EnvVarSourceArgs{
										SecretKeyRef: corev1.SecretKeySelectorArgs{
											Name: pulumi.String(fmt.Sprintf("%s-admin-key", args.Name)),
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
							VolumeMounts: corev1.VolumeMountArray{
								&corev1.VolumeMountArgs{
									Name:      pulumi.String("postgres-certs"),
									MountPath: pulumi.String("/postgres-certs"),
								},
								&corev1.VolumeMountArgs{
									Name:      pulumi.String("redis-certs"),
									MountPath: pulumi.String("/redis-certs"),
								},
							},
						},
					},
					TerminationGracePeriodSeconds: pulumi.Int(30),
				},
			},
		},
	}, pulumi.DependsOn([]pulumi.Resource{args.DbCert, args.RedisCert, streamSecret, adminKey}))
	if err != nil {
		return err
	}

	return nil
}

func deployRbac(ctx *pulumi.Context, name string) error {
	_, err := corev1.NewServiceAccount(ctx, fmt.Sprintf("%s-sa", name), &corev1.ServiceAccountArgs{
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
	_, err = rbacv1.NewRole(ctx, fmt.Sprintf("%s-role", name), &rbacv1.RoleArgs{
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
	_, err = rbacv1.NewRoleBinding(ctx, fmt.Sprintf("%s-rb", name), &rbacv1.RoleBindingArgs{
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
				Namespace: pulumi.String("default"),
			},
		},
	})
	if err != nil {
		return err
	}
	return nil
}

func deployService(ctx *pulumi.Context, name string) error {
	_, err := corev1.NewService(ctx, fmt.Sprintf("%s-service", name), &corev1.ServiceArgs{
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
					Port:       pulumi.Int(80),
					TargetPort: pulumi.Int(3001),
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

	_, err = corev1.NewService(ctx, fmt.Sprintf("%s-connector-service", name), &corev1.ServiceArgs{
		ApiVersion: pulumi.String("v1"),
		Kind:       pulumi.String("Service"),
		Metadata: &metav1.ObjectMetaArgs{
			Name: pulumi.String(fmt.Sprintf("%s-connector", name)),
			Labels: pulumi.StringMap{
				"app": pulumi.String(fmt.Sprintf("%s-connector", name)),
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
				"app": pulumi.String(name),
			},
		},
	})
	if err != nil {
		return err
	}

	return nil
}

func deployIngress(ctx *pulumi.Context, hn string, name string) error {
	err := ingress.DeployMapping(ctx, &ingress.MappingArgs{
		Name:     name,
		Hostname: hn,
		Prefix:   "/",
		Rewrite:  "/",
		Service:  name,
	})
	if err != nil {
		return err
	}

	return nil
}

func deployPlaygroundIngress(ctx *pulumi.Context, hn string, name string) error {
	err := ingress.DeployMapping(ctx, &ingress.MappingArgs{
		Name:     fmt.Sprintf("%s-graphql", name),
		Hostname: hn,
		Prefix:   "/graphql",
		Rewrite:  "/graphql",
		Service:  name,
	})
	if err != nil {
		return err
	}

	return nil
}
