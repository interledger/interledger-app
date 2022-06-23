package main

import (
	"github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes"
	v1 "github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes/core/v1"
	"github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes/helm/v3"
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
				Name: pulumi.String("vault"),
			},
		}, pulumi.Provider(kubeProvider))

		_, err = helm.NewRelease(ctx, "vault", &helm.ReleaseArgs{
			Version:   pulumi.String("0.20.1"),
			Chart:     pulumi.String("vault"),
			Namespace: namespace.Metadata.Name(),
			Name:      pulumi.String("vault"),
			RepositoryOpts: &helm.RepositoryOptsArgs{
				Repo: pulumi.String("https://helm.releases.hashicorp.com"),
			},
			Values: pulumi.Map{
				"global": pulumi.Map{
					"externalVaultAddr": pulumi.String("https://vault1.fynbos.cloud:8200"),
				},
			},
		}, pulumi.Provider(kubeProvider), pulumi.DependsOn([]pulumi.Resource{namespace}))

		if err != nil {
			return err
		}

		return nil
	})
}
