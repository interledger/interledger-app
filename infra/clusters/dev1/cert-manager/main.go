package main

import (
	"fmt"
	"github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes"
	"github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes/helm/v3"
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

		_, err = helm.NewRelease(ctx, "cert-manager", &helm.ReleaseArgs{
			Version:         pulumi.String("1.8.0"),
			Chart:           pulumi.String("cert-manager"),
			Namespace:       pulumi.String("cert-manager"),
			CreateNamespace: pulumi.BoolPtr(true),
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
			},
		}, pulumi.Provider(kubeProvider))

		if err != nil {
			return err
		}

		fmt.Println(kubeProvider.ID())
		return nil
	})
}
