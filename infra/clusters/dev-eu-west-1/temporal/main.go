package main

import (
	"github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes"
	v1 "github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes/core/v1"
	metav1 "github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes/meta/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"
	"gitlab.com/fynbos/infra/services/temporal"
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
		fynbosConfig := config.New(ctx, "fynbos")
		ecrRepo := fynbosConfig.Get("ecrRepo")

		namespace, err := v1.NewNamespace(ctx, "temporal", &v1.NamespaceArgs{
			Metadata: metav1.ObjectMetaArgs{
				Name: pulumi.String("temporal"),
			},
		}, pulumi.Provider(kubeProvider))
		if err != nil {
			return err
		}

		err = temporal.DeployTemporalDev(ctx, ecrRepo, "latest", namespace.Metadata.Name(),
			pulumi.Provider(kubeProvider), pulumi.DependsOn([]pulumi.Resource{namespace}))
		if err != nil {
			return err
		}

		return nil
	})
}
