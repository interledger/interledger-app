package main

import (
	"github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes"
	corev1 "github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes/core/v1"
	"github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes/helm/v3"
	metav1 "github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes/meta/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"
	k8s "gitlab.com/fynbos/infra/aws/modules/kubernetes"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		conf := config.New(ctx, "cluster")

		oidcProvider := conf.Get("oidcProvider")
		clusterName := conf.Get("name")

		fynbosConf := config.New(ctx, "fynbos")
		accountId := fynbosConf.Get("accountId")
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

		//Install
		_, err = helm.NewChart(ctx, "cilium", helm.ChartArgs{
			FetchArgs: &helm.FetchArgs{
				Repo: pulumi.String("https://helm.cilium.io/"),
			},
			Namespace: pulumi.String("kube-system"),
			Chart:     pulumi.String("cilium"),
			Version:   pulumi.String("1.12.0"),
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

		baselineStack, err := pulumi.NewStackReference(ctx, "fynbos/aws-dev-baseline/main", nil)
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

		// Setup AWS LB Controller (https://docs.aws.amazon.com/eks/latest/userguide/aws-load-balancer-controller.html)
		lbRole, err := k8s.CreateLbControllerRole(ctx, k8s.CreateLbControllerRoleArgs{
			AccountId:      pulumi.String(accountId),
			OidcProvider:   pulumi.String(oidcProvider),
			Namespace:      pulumi.String("kube-system"),
			ServiceAccount: pulumi.String("aws-load-balancer-controller"),
		})
		if err != nil {
			return err
		}

		serviceAccount, err := corev1.NewServiceAccount(ctx, "aws-load-balancer-controller", &corev1.ServiceAccountArgs{
			ApiVersion: pulumi.String("v1"),
			Kind:       pulumi.String("ServiceAccount"),
			Metadata: &metav1.ObjectMetaArgs{
				Name:      pulumi.String("aws-load-balancer-controller"),
				Namespace: pulumi.String("kube-system"),
				Annotations: pulumi.StringMap{
					"eks.amazonaws.com/role-arn": lbRole.Arn,
				},
			},
		}, pulumi.Provider(kubeProvider), pulumi.DependsOn([]pulumi.Resource{lbRole}))
		if err != nil {
			return err
		}

		_, err = helm.NewChart(ctx, "aws-load-balancer-controller", helm.ChartArgs{
			FetchArgs: &helm.FetchArgs{
				Repo: pulumi.String("https://aws.github.io/eks-charts"),
			},
			Namespace: pulumi.String("kube-system"),
			Chart:     pulumi.String("aws-load-balancer-controller"),
			Version:   pulumi.String("1.4.2"),
			Values: pulumi.Map{
				"image": pulumi.Map{
					"repository": pulumi.String("602401143452.dkr.ecr.eu-west-1.amazonaws.com/amazon/aws-load-balancer-controller"),
				},
				"clusterName": pulumi.String(clusterName),
				"serviceAccount": pulumi.Map{
					"name":   pulumi.String("aws-load-balancer-controller"),
					"create": pulumi.Bool(false),
				},
				"hostNetwork": pulumi.Bool(true), // NB Required as we are using Cilium CNI
			},
		}, pulumi.Provider(kubeProvider), pulumi.DependsOn([]pulumi.Resource{serviceAccount, kubeProvider}))
		if err != nil {
			return err
		}

		return nil
	})
}
