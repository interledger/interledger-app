package main

import (
	"github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes"
	corev1 "github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes/core/v1"
	"github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes/helm/v3"
	metav1 "github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes/meta/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
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

		namespace, err := corev1.NewNamespace(ctx, "vault", &corev1.NamespaceArgs{
			Metadata: metav1.ObjectMetaArgs{
				Name: pulumi.String("vault"),
			},
		}, pulumi.Provider(kubeProvider))
		if err != nil {
			return err
		}

		_, err = helm.NewChart(ctx, "vault", helm.ChartArgs{
			Namespace: namespace.Metadata.Name().Elem(),
			Chart:     pulumi.String("vault"),
			Version:   pulumi.String("0.20.0"),
			FetchArgs: helm.FetchArgs{
				Repo: pulumi.String("https://helm.releases.hashicorp.com"),
			},
			Values: pulumi.Map{
				"server": pulumi.Map{
					"nodeSelector": pulumi.StringMap{
						"vault_in_k8s": pulumi.String("true"),
					},
					"tolerations": pulumi.MapArray{
						pulumi.Map{
							"key":      pulumi.String("taint_for_consul_xor_vault"),
							"operator": pulumi.String("Equal"),
							"value":    pulumi.String("true"),
							"effect":   pulumi.String("NoExecute"),
						},
					},
					"ha": pulumi.Map{
						"enabled": pulumi.Bool(true),
						"raft": pulumi.Map{
							"enabled": pulumi.Bool(true),
						},
					},
				},
			},
		}, pulumi.Provider(kubeProvider), pulumi.DependsOn([]pulumi.Resource{namespace}))

		return err
	})
}
