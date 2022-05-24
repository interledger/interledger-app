package main

import (
	"github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes"
	"github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes/helm/v3"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"
	k8s "gitlab.com/fynbos/infra/aws/modules/kubernetes"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		conf := config.New(ctx, "cluster")
		// TODO: figure out how to extract oidcID from OidcProviderOutput exported in kubernetes stack.
		oidcProvider := conf.Get("oidcProvider")
		clusterName := conf.Get("name")

		fynbosConf := config.New(ctx, "fynbos")
		accountId := fynbosConf.Get("accountId")
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

		//Install
		_, err = helm.NewChart(ctx, "cilium", helm.ChartArgs{
			FetchArgs: &helm.FetchArgs{
				Repo: pulumi.String("https://helm.cilium.io/"),
			},
			Namespace: pulumi.String("kube-system"),
			Chart:     pulumi.String("cilium"),
			Version:   pulumi.String("1.11.5"),
			Values: pulumi.Map{
				"egressMasqueradeInterfaces": pulumi.String("eth0"),
			},
		}, pulumi.Provider(kubeProvider))
		if err != nil {
			return err
		}

		err = k8s.ConfigureClusterRolesAndPsp(ctx, pulumi.Provider(kubeProvider))
		if err != nil {
			return err
		}

		baselineStack, err := pulumi.NewStackReference(ctx, "fynbos/aws-shared-baseline/main", nil)
		if err != nil {
			return err
		}
		ebsKmsKeyArn := baselineStack.GetStringOutput(pulumi.String("ebsEncryptionKeyArn"))

		err = k8s.DeployEbsCsi(ctx, k8s.EbsCsiArgs{
			ClusterName:  clusterName,
			EbsKmsKeyArn: ebsKmsKeyArn,
			OidcProvider: oidcProvider,
			AccountId:    accountId,
		})
		if err != nil {
			return err
		}

		err = k8s.DeployDefaultCSIStorageClass(ctx, pulumi.Provider(kubeProvider))
		if err != nil {
			return err
		}

		return nil
	})
}
