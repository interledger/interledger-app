package main

import (
	"github.com/pulumi/pulumi-aws/sdk/v4/go/aws/iam"
	"github.com/pulumi/pulumi-aws/sdk/v4/go/aws/kms"
	"github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes"
	"github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes/apiextensions"
	appsv1 "github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes/apps/v1"
	corev1 "github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes/core/v1"
	v1 "github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes/core/v1"
	"github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes/helm/v3"
	metav1 "github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes/meta/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"
	k8s "gitlab.com/fynbos/infra/aws/modules/kubernetes"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		k8sStack, err := pulumi.NewStackReference(ctx, "fynbos/aws-shared-eu-west-1-shared-k8s/main", nil)
		if err != nil {
			return err
		}
		kubeConfig := k8sStack.GetStringOutput(pulumi.String("kubeconfig"))
		kubeProvider, err := kubernetes.NewProvider(ctx, "kubernetes-provider", &kubernetes.ProviderArgs{
			Kubeconfig: kubeConfig,
		})
		if err != nil {
			return err
		}
		oidcProvider := k8sStack.GetStringOutput(pulumi.String("oidcProvider"))
		cfg := config.New(ctx, "fynbos")
		accountID := cfg.Require("accountId")

		role, err := createRole(ctx, CreateRoleArgs{
			AccountId:      pulumi.String(accountID),
			OidcProvider:   oidcProvider,
			Namespace:      pulumi.String("cert-manager"),
			ServiceAccount: pulumi.String("cert-manager"),
		})
		if err != nil {
			return err
		}

		namespace, err := v1.NewNamespace(ctx, "namespace", &v1.NamespaceArgs{
			Metadata: metav1.ObjectMetaArgs{
				Name: pulumi.String("cert-manager"),
				Labels: pulumi.StringMap{
					"kubernetes.io/metadata.name":        pulumi.String("cert-manager"),
					"name":                               pulumi.String("cert-manager"),
					"pod-security.kubernetes.io/enforce": pulumi.String("privileged"),
					"pod-security.kubernetes.io/audit":   pulumi.String("privileged"),
					"pod-security.kubernetes.io/warn":    pulumi.String("privileged"),
				},
			},
			Spec: v1.NamespaceSpecArgs{
				Finalizers: pulumi.StringArray{pulumi.String("kubernetes")},
			},
		}, pulumi.Import(pulumi.ID("cert-manager")), pulumi.Provider(kubeProvider))
		if err != nil {
			return err
		}

		release, err := helm.NewRelease(ctx, "cert-manager", &helm.ReleaseArgs{
			Version:         pulumi.String("1.8.0"),
			Chart:           pulumi.String("cert-manager"),
			Namespace:       namespace.Metadata.Name(),
			CreateNamespace: pulumi.BoolPtr(false),
			RepositoryOpts: &helm.RepositoryOptsArgs{
				Repo: pulumi.String("https://charts.jetstack.io"),
			},
			Values: pulumi.Map{
				"installCRDs": pulumi.Bool(true),
				// required https://cert-manager.io/docs/installation/compatibility/#aws-eks
				"webhook": pulumi.Map{
					"hostNetwork": pulumi.BoolPtr(true),
					"securePort":  pulumi.String("10260"), // required https://github.com/cert-manager/website/issues/403
				},
				"serviceAccount": pulumi.Map{
					"name": pulumi.String("cert-manager"),
					"annotations": pulumi.Map{
						"eks.amazonaws.com/role-arn": role.Arn,
					},
				},
				"securityContext": pulumi.Map{
					"enabled": pulumi.Bool(true),
					"fsGroup": pulumi.Int(1001),
				},
				"extraArgs": pulumi.StringArray{
					pulumi.String("--dns01-recursive-nameservers=\"cloudflared.default:53\""),
					pulumi.String("--dns01-recursive-nameservers-only"),
				},
			},
		}, pulumi.Provider(kubeProvider), pulumi.DependsOn([]pulumi.Resource{role}))
		if err != nil {
			return err
		}

		err = deployDnsOverHttpsProxy(ctx, kubeProvider)
		if err != nil {
			return err
		}

		_, err = apiextensions.NewCustomResource(ctx, "fynbos-cloud-issuer", &apiextensions.CustomResourceArgs{
			ApiVersion: pulumi.String("cert-manager.io/v1"),
			Kind:       pulumi.String("ClusterIssuer"),
			Metadata: metav1.ObjectMetaArgs{
				Name:      pulumi.String("fynbos-cloud-issuer"),
				Namespace: pulumi.String("cert-manager"),
			},
			OtherFields: kubernetes.UntypedArgs{
				"spec": pulumi.Map{
					"acme": pulumi.Map{
						"server": pulumi.String("https://acme-v02.api.letsencrypt.org/directory"),
						"email":  pulumi.String("matt@fynbos.dev"),
						"privateKeySecretRef": pulumi.Map{
							"name": pulumi.String("fynbos-issuer-account-key"),
						},
						"solvers": pulumi.MapArray{
							pulumi.Map{
								"selector": pulumi.Map{
									"dnsZones": pulumi.StringArray{
										pulumi.String("fynbos.cloud"),
									},
								},
								"dns01": pulumi.Map{
									"route53": pulumi.Map{
										"hostedZoneID": pulumi.String("Z09140632GCY8MFP051HD"),
										"region":       pulumi.String("eu-west-1"),
									},
								},
							},
						},
					},
				},
			},
		}, pulumi.Provider(kubeProvider), pulumi.DependsOn([]pulumi.Resource{release, role}))

		// Create KMS Key
		kmsIssuerKey, err := kms.NewKey(ctx, "kms-issuer-key", &kms.KeyArgs{
			DeletionWindowInDays:  pulumi.Int(30),
			Description:           pulumi.String("KMS Issuer Key"),
			CustomerMasterKeySpec: pulumi.String("RSA_2048"),
			KeyUsage:              pulumi.String("SIGN_VERIFY"),
		})
		if err != nil {
			return err
		}

		kmsIssuerAlias, err := kms.NewAlias(ctx, "kms-issuer-alias", &kms.AliasArgs{
			TargetKeyId: kmsIssuerKey.KeyId,
			Name:        pulumi.String("alias/kms-issuer"),
		}, pulumi.DependsOn([]pulumi.Resource{kmsIssuerKey}))

		// Create KMS Role
		kmsRole, err := createKmsIssuerRole(ctx, CreateKmsIssuerRole{
			AccountId:      pulumi.String(accountID),
			OidcProvider:   oidcProvider,
			Namespace:      pulumi.String("cert-manager"),
			ServiceAccount: pulumi.String("kms-issuer"),
			KmsKeyArn:      kmsIssuerKey.Arn,
		})
		if err != nil {
			return err
		}

		_, err = helm.NewRelease(ctx, "kms-issuer", &helm.ReleaseArgs{
			Version:   pulumi.String("1.0.1"),
			Chart:     pulumi.String("kms-issuer"),
			Namespace: pulumi.String("cert-manager"),
			RepositoryOpts: &helm.RepositoryOptsArgs{
				Repo: pulumi.String("https://skyscanner.github.io/kms-issuer"),
			},
			Values: pulumi.Map{
				"serviceAccount": pulumi.Map{
					"name": pulumi.String("kms-issuer"),
					"annotations": pulumi.Map{
						"eks.amazonaws.com/role-arn": kmsRole.Arn,
					},
				},
				"env": pulumi.MapArray{
					pulumi.Map{
						"name":  pulumi.String("AWS_REGION"),
						"value": pulumi.String("eu-west-1"),
					},
				},
			},
		}, pulumi.Provider(kubeProvider), pulumi.DependsOn([]pulumi.Resource{kmsRole, kmsIssuerAlias}))
		if err != nil {
			return err
		}

		_, err = helm.NewRelease(ctx, "csi-driver-cert-manager", &helm.ReleaseArgs{
			Version:   pulumi.String("0.3.2"),
			Chart:     pulumi.String("cert-manager-csi-driver"),
			Namespace: pulumi.String("cert-manager"),
			RepositoryOpts: &helm.RepositoryOptsArgs{
				Repo: pulumi.String("https://charts.jetstack.io"),
			},
			Values: pulumi.Map{},
		}, pulumi.Provider(kubeProvider), pulumi.DependsOn([]pulumi.Resource{release}))
		if err != nil {
			return err
		}

		return nil
	})
}

type CreateRoleArgs struct {
	AccountId      pulumi.StringPtrInput
	OidcProvider   pulumi.StringPtrInput
	Namespace      pulumi.StringPtrInput
	ServiceAccount pulumi.StringPtrInput
}

func createRole(ctx *pulumi.Context, args CreateRoleArgs) (*iam.Role, error) {
	policy, err := iam.GetPolicyDocument(ctx, &iam.GetPolicyDocumentArgs{
		Statements: []iam.GetPolicyDocumentStatement{
			{
				Effect: stringPtr("Allow"),
				Actions: []string{
					"route53:GetChange",
				},
				Resources: []string{
					"arn:aws:route53:::change/*",
				},
			},
			{
				Effect: stringPtr("Allow"),
				Actions: []string{
					"route53:ChangeResourceRecordSets",
					"route53:ListResourceRecordSets",
				},
				Resources: []string{
					"arn:aws:route53:::hostedzone/*",
				},
			},
			{
				Effect: stringPtr("Allow"),
				Actions: []string{
					"route53:ListHostedZonesByName",
				},
				Resources: []string{
					"*",
				},
			},
		},
	})
	if err != nil {
		return nil, err
	}

	trustPolicy := k8s.NewIamTrustPolicyDocumentV2(ctx, args.AccountId, args.OidcProvider, args.Namespace, args.ServiceAccount)

	role, err := iam.NewRole(ctx, "cert-manager-role", &iam.RoleArgs{
		Name:        pulumi.String("eks-shared-cert-manager-role"),
		Description: pulumi.String("AWS role for cert manager within the shared EKS cluster"),
		InlinePolicies: iam.RoleInlinePolicyArray{
			iam.RoleInlinePolicyArgs{
				Name:   pulumi.String("route53"),
				Policy: pulumi.String(policy.Json),
			},
		},
		AssumeRolePolicy: trustPolicy,
	})

	return role, err
}

func stringPtr(s string) *string {
	return &s
}

func deployDnsOverHttpsProxy(ctx *pulumi.Context, provider pulumi.ProviderResource) error {
	_, err := appsv1.NewDeployment(ctx, "cloudflared-deployment", &appsv1.DeploymentArgs{
		ApiVersion: pulumi.String("apps/v1"),
		Kind:       pulumi.String("Deployment"),
		Metadata: &metav1.ObjectMetaArgs{
			Name: pulumi.String("cloudflared"),
			Labels: pulumi.StringMap{
				"app": pulumi.String("cloudflared"),
			},
		},
		Spec: &appsv1.DeploymentSpecArgs{
			Replicas: pulumi.Int(1),
			Selector: &metav1.LabelSelectorArgs{
				MatchLabels: pulumi.StringMap{
					"app": pulumi.String("cloudflared"),
				},
			},
			Template: &corev1.PodTemplateSpecArgs{
				Metadata: &metav1.ObjectMetaArgs{
					Labels: pulumi.StringMap{
						"app": pulumi.String("cloudflared"),
					},
				},
				Spec: &corev1.PodSpecArgs{
					Containers: corev1.ContainerArray{
						&corev1.ContainerArgs{
							Name:  pulumi.String("cloudflared"),
							Image: pulumi.String("cloudflare/cloudflared"),
							Ports: corev1.ContainerPortArray{
								&corev1.ContainerPortArgs{
									ContainerPort: pulumi.Int(5553),
									Protocol:      pulumi.String("UDP"),
								},
							},
							Command: pulumi.StringArray{
								pulumi.String("cloudflared"),
							},
							Args: pulumi.StringArray{
								pulumi.String("proxy-dns"),
							},
							Env: corev1.EnvVarArray{
								corev1.EnvVarArgs{
									Name:  pulumi.String("TUNNEL_DNS_PORT"),
									Value: pulumi.String("5553"),
								},
								corev1.EnvVarArgs{
									Name:  pulumi.String("TUNNEL_DNS_ADDRESS"),
									Value: pulumi.String("0.0.0.0"),
								},
							},
						},
					},
				},
			},
		},
	}, pulumi.Provider(provider))
	if err != nil {
		return err
	}

	_, err = corev1.NewService(ctx, "cloudflared-service", &corev1.ServiceArgs{
		ApiVersion: pulumi.String("v1"),
		Kind:       pulumi.String("Service"),
		Metadata: &metav1.ObjectMetaArgs{
			Name: pulumi.String("cloudflared"),
		},
		Spec: &corev1.ServiceSpecArgs{
			Selector: pulumi.StringMap{
				"app": pulumi.String("cloudflared"),
			},
			Ports: corev1.ServicePortArray{
				&corev1.ServicePortArgs{
					Protocol:   pulumi.String("UDP"),
					Port:       pulumi.Int(53),
					TargetPort: pulumi.Int(5553),
				},
			},
		},
	}, pulumi.Provider(provider))
	if err != nil {
		return err
	}

	return nil
}

type CreateKmsIssuerRole struct {
	AccountId      pulumi.StringPtrInput
	OidcProvider   pulumi.StringPtrInput
	Namespace      pulumi.StringPtrInput
	ServiceAccount pulumi.StringPtrInput
	KmsKeyArn      pulumi.StringPtrInput
}

func createKmsIssuerRole(ctx *pulumi.Context, args CreateKmsIssuerRole) (*iam.Role, error) {
	policy := pulumi.All(args.KmsKeyArn).ApplyT(func(args []interface{}) (string, error) {
		keyArn := args[0].(string)

		kmsAccess, err := iam.GetPolicyDocument(ctx, &iam.GetPolicyDocumentArgs{
			Statements: []iam.GetPolicyDocumentStatement{
				{
					Effect: stringPtr("Allow"),
					Actions: []string{
						"kms:DescribeKey",
						"kms:Sign",
						"kms:Verify",
						"kms:GetPublicKey",
					},
					Resources: []string{
						keyArn,
					},
				},
			},
		})
		return kmsAccess.Json, err
	}).(pulumi.StringOutput)

	trustPolicy := k8s.NewIamTrustPolicyDocumentV2(ctx, args.AccountId, args.OidcProvider, args.Namespace, args.ServiceAccount)

	role, err := iam.NewRole(ctx, "kms-issuer-role", &iam.RoleArgs{
		Name:        pulumi.String("eks-shared-kms-issuer-role"),
		Description: pulumi.String("AWS role for kms issuer within the shared EKS cluster"),
		InlinePolicies: iam.RoleInlinePolicyArray{
			iam.RoleInlinePolicyArgs{
				Name:   pulumi.String("kms"),
				Policy: policy,
			},
		},
		AssumeRolePolicy: trustPolicy,
	})

	return role, err
}
