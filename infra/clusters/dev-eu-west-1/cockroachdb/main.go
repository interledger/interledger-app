package main

import (
	"github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes"
	"github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes/apiextensions"
	v1 "github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes/core/v1"
	metav1 "github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes/meta/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
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
		// Create KMSIssuer for internal Vault TLS
		_, err = apiextensions.NewCustomResource(ctx, "crdb-issuer", &apiextensions.CustomResourceArgs{
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

		//// Should this issuer be here or in cert-manager?
		//_, err = helm.NewRelease(ctx, "cert-manager", &helm.ReleaseArgs{
		//	Namespace: namespace.Metadata.Name().Elem(),
		//	Chart:     pulumi.String("vault"),
		//	Version:   pulumi.String("0.20.0"),
		//	RepositoryOpts: helm.RepositoryOptsArgs{
		//		Repo: pulumi.String("https://helm.releases.hashicorp.com"),
		//	},
		//	Values: pulumi.Map{},
		//}, pulumi.Provider(kubeProvider))
		//
		//if err != nil {
		//	return err
		//}

		return nil
	})
}
