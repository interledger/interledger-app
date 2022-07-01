package main

import (
	"github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes"
	"github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes/apiextensions"
	v1 "github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes/core/v1"
	metav1 "github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes/meta/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"gitlab.com/fynbos/infra/services/kratos/v1"
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

		namespace, err := v1.NewNamespace(ctx, "kratos", &v1.NamespaceArgs{
			Metadata: metav1.ObjectMetaArgs{
				Name: pulumi.String("kratos"),
			},
		}, pulumi.Provider(kubeProvider))
		if err != nil {
			return err
		}

		// Create service account
		sa, err := v1.NewServiceAccount(ctx, "kratos-sa", &v1.ServiceAccountArgs{
			ApiVersion: pulumi.String("v1"),
			Kind:       pulumi.String("ServiceAccount"),
			Metadata: metav1.ObjectMetaArgs{
				Name:      pulumi.String("kratos"),
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
						"path":   pulumi.String("pki/dev-int/sign/crdb-client"),
						"server": pulumi.String("https://vault1.fynbos.cloud:8200"),
						"auth": pulumi.Map{
							"kubernetes": pulumi.Map{
								"role":      pulumi.String("k8s-app"),
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

		certSecretName := "crdb-kratos-cert"
		cert, err := apiextensions.NewCustomResource(ctx, "crdb-cert", &apiextensions.CustomResourceArgs{
			ApiVersion: pulumi.String("cert-manager.io/v1"),
			Kind:       pulumi.String("Certificate"),
			Metadata: metav1.ObjectMetaArgs{
				Name:      pulumi.String("crdb-kratos-cert"),
				Namespace: namespace.Metadata.Name(),
			},
			OtherFields: kubernetes.UntypedArgs{
				"spec": pulumi.Map{
					"secretName":  pulumi.String(certSecretName),
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
					"commonName": pulumi.String("kratos"),
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

		chart, err := kratos.DeployKratos(ctx, kratos.DeployKratosArgs{
			Domain:             "https://eu1.fynbos.dev",
			DefaultSecret:      "CHANGE-ME-I-AM-VERY-INSECURE1234",
			CertSecretName:     pulumi.String(certSecretName),
			Namespace:          namespace.Metadata.Name().Elem(),
			ServiceAccountName: sa.Metadata.Name(),
		}, pulumi.Provider(kubeProvider), pulumi.DependsOn([]pulumi.Resource{cert, namespace}))
		if err != nil {
			return err
		}

		err = kratos.DeployIngress(ctx, kratos.DeployIngressArgs{
			Hostname:  pulumi.String("eu1.fynbos.dev"),
			Namespace: namespace.Metadata.Name().Elem(),
		}, pulumi.Provider(kubeProvider), pulumi.DependsOn([]pulumi.Resource{chart, namespace}))
		if err != nil {
			return err
		}

		return nil
	})
}
