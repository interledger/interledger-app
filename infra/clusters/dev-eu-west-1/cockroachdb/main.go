package main

import (
	"github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes"
	"github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes/apiextensions"
	v1 "github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes/core/v1"
	metav1 "github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes/meta/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	crdb "gitlab.com/fynbos/infra/services/cockroach/v1"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		k8sStack, err := pulumi.NewStackReference(ctx, "fynbos/aws-dev-eu-west-1-dev-k8s/main", nil)
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

		namespace, err := v1.NewNamespace(ctx, "namespace", &v1.NamespaceArgs{
			Metadata: metav1.ObjectMetaArgs{
				Name: pulumi.String("cockroachdb"),
			},
		}, pulumi.Provider(kubeProvider))
		if err != nil {
			return err
		}

		// Create service account
		sa, err := v1.NewServiceAccount(ctx, "issuer-sa", &v1.ServiceAccountArgs{
			ApiVersion: pulumi.String("v1"),
			Kind:       pulumi.String("ServiceAccount"),
			Metadata: metav1.ObjectMetaArgs{
				Name:      pulumi.String("crdb-issuer"),
				Namespace: namespace.Metadata.Name(),
			},
		}, pulumi.Provider(kubeProvider), pulumi.DependsOn([]pulumi.Resource{namespace}))
		if err != nil {
			return err
		}

		// Create node issuer - needs a role from vault
		issuer, err := apiextensions.NewCustomResource(ctx, "crdb-issuer", &apiextensions.CustomResourceArgs{
			ApiVersion: pulumi.String("cert-manager.io/v1"),
			Kind:       pulumi.String("Issuer"),
			Metadata: metav1.ObjectMetaArgs{
				Name:      pulumi.String("crdb-issuer"),
				Namespace: namespace.Metadata.Name(),
			},
			OtherFields: kubernetes.UntypedArgs{
				"spec": pulumi.Map{
					"vault": pulumi.Map{
						"path":   pulumi.String("pki/dev-int/sign/crdb-node"),
						"server": pulumi.String("https://vault1.fynbos.cloud:8200"),
						"auth": pulumi.Map{
							"kubernetes": pulumi.Map{
								"role":      pulumi.String("crdb-node"),
								"mountPath": pulumi.String("/v1/auth/k8s-dev-euw1"),
								"secretRef": pulumi.Map{
									"name": sa.Secrets.Index(pulumi.Int(0)).Name(),
									"key":  pulumi.String("token"),
								},
							},
						},
					},
				},
			},
		}, pulumi.Provider(kubeProvider), pulumi.DependsOn([]pulumi.Resource{namespace, sa}))
		if err != nil {
			return err
		}

		// Create Cert
		nodeCert, err := apiextensions.NewCustomResource(ctx, "crdb-tls-cert", &apiextensions.CustomResourceArgs{
			ApiVersion: pulumi.String("cert-manager.io/v1"),
			Kind:       pulumi.String("Certificate"),
			Metadata: metav1.ObjectMetaArgs{
				Name:      pulumi.String("crdb-tls-cert"),
				Namespace: namespace.Metadata.Name(),
			},
			OtherFields: kubernetes.UntypedArgs{
				"spec": pulumi.Map{
					"secretName":  pulumi.String("cockroachdb-node"),
					"duration":    pulumi.String("720h"),
					"renewBefore": pulumi.String("168h"),
					"usages": pulumi.StringArray{
						pulumi.String("digital signature"),
						pulumi.String("key encipherment"),
						pulumi.String("server auth"),
						pulumi.String("client auth"),
					},
					"privateKey": pulumi.Map{
						"algorithm": pulumi.String("RSA"),
						"size":      pulumi.Int(2048),
					},
					"commonName": pulumi.String("node"),
					"subject": pulumi.Map{
						"organizations": pulumi.StringArray{
							pulumi.String("Cockroach"),
						},
					},
					"dnsNames": pulumi.StringArray{
						pulumi.String("localhost"),
						pulumi.String("127.0.0.1"),
						pulumi.Sprintf("%s-public", "cockroachdb"),
						pulumi.Sprintf("%s-public.%s", "cockroachdb", namespace.Metadata.Name().Elem()),
						pulumi.Sprintf("%s-public.%s.svc.cluster.local", "cockroachdb", namespace.Metadata.Name().Elem()),
						pulumi.Sprintf("*.%s", "cockroachdb"),
						pulumi.Sprintf("*.%s.%s", "cockroachdb", namespace.Metadata.Name().Elem()),
						pulumi.Sprintf("*.%s.%s.svc.cluster.local", "cockroachdb", namespace.Metadata.Name().Elem()),
					},
					"issuerRef": pulumi.Map{
						"name":  issuer.Metadata.Name(),
						"kind":  issuer.Kind,
						"group": pulumi.String("cert-manager.io"),
					},
				},
			},
		}, pulumi.Provider(kubeProvider), pulumi.DependsOn([]pulumi.Resource{namespace, issuer}))
		if err != nil {
			return err
		}

		// Create Cert
		rootCert, err := apiextensions.NewCustomResource(ctx, "crdb-root-cert", &apiextensions.CustomResourceArgs{
			ApiVersion: pulumi.String("cert-manager.io/v1"),
			Kind:       pulumi.String("Certificate"),
			Metadata: metav1.ObjectMetaArgs{
				Name:      pulumi.String("crdb-root-cert"),
				Namespace: namespace.Metadata.Name(),
			},
			OtherFields: kubernetes.UntypedArgs{
				"spec": pulumi.Map{
					"secretName":  pulumi.String("cockroachdb-root"),
					"duration":    pulumi.String("168h"),
					"renewBefore": pulumi.String("24h"),
					"usages": pulumi.StringArray{
						pulumi.String("digital signature"),
						pulumi.String("key encipherment"),
						pulumi.String("client auth"),
					},
					"privateKey": pulumi.Map{
						"algorithm": pulumi.String("RSA"),
						"size":      pulumi.Int(2048),
					},
					"commonName": pulumi.String("root"),
					"subject": pulumi.Map{
						"organizations": pulumi.StringArray{
							pulumi.String("Cockroach"),
						},
					},
					"issuerRef": pulumi.Map{
						"name":  issuer.Metadata.Name(),
						"kind":  issuer.Kind,
						"group": pulumi.String("cert-manager.io"),
					},
				},
			},
		}, pulumi.Provider(kubeProvider), pulumi.DependsOn([]pulumi.Resource{namespace, issuer}))
		if err != nil {
			return err
		}

		serviceAccount, err := crdb.DeployRbac(ctx, namespace.Metadata.Name().Elem(), pulumi.Provider(kubeProvider), pulumi.DependsOn([]pulumi.Resource{namespace}))
		if err != nil {
			return err
		}

		_, err = crdb.DeployStatefulSet(ctx, crdb.StatefulSetArgs{
			Namespace:          namespace.Metadata.Name().Elem(),
			Replicas:           3,
			ServiceAccountName: serviceAccount.Metadata.Name(),
			CertSecretName:     pulumi.String("cockroachdb-node"),
			ClientSecretName:   pulumi.String("cockroachdb-root"),
		}, pulumi.Provider(kubeProvider), pulumi.DependsOn([]pulumi.Resource{serviceAccount, namespace, nodeCert, rootCert}))
		if err != nil {
			return err
		}

		err = crdb.DeployPublicService(ctx, namespace.Metadata.Name(), pulumi.Provider(kubeProvider), pulumi.DependsOn([]pulumi.Resource{namespace}))
		if err != nil {
			return err
		}
		err = crdb.DeployPrivateService(ctx, namespace.Metadata.Name(), pulumi.Provider(kubeProvider), pulumi.DependsOn([]pulumi.Resource{namespace}))
		if err != nil {
			return err
		}

		err = crdb.DeployPodDistributionBudget(ctx, pulumi.Provider(kubeProvider))
		if err != nil {
			return err
		}

		return nil
	})
}
