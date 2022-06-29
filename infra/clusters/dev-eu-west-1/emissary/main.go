package main

import (
	"github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes"
	"github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes/apiextensions"
	v1 "github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes/core/v1"
	"github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes/helm/v3"
	metav1 "github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes/meta/v1"
	"github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes/yaml"
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

		crds, err := yaml.NewConfigFile(ctx, "crd-config",
			&yaml.ConfigFileArgs{
				File: "./emissary-crds-3.0.yaml",
			}, pulumi.Provider(kubeProvider))
		if err != nil {
			return err
		}

		namespace, err := v1.NewNamespace(ctx, "namespace", &v1.NamespaceArgs{
			Metadata: metav1.ObjectMetaArgs{
				Name: pulumi.String("emissary"),
			},
		}, pulumi.Provider(kubeProvider))
		if err != nil {
			return err
		}

		// Create service account
		sa, err := v1.NewServiceAccount(ctx, "emissary-sa", &v1.ServiceAccountArgs{
			ApiVersion: pulumi.String("v1"),
			Kind:       pulumi.String("ServiceAccount"),
			Metadata: metav1.ObjectMetaArgs{
				Name:      pulumi.String("emissary"),
				Namespace: namespace.Metadata.Name(),
			},
		}, pulumi.Provider(kubeProvider), pulumi.DependsOn([]pulumi.Resource{namespace}))
		if err != nil {
			return err
		}

		_, err = helm.NewRelease(ctx, "emissary", &helm.ReleaseArgs{
			Version:         pulumi.String("8.0.0"),
			Chart:           pulumi.String("emissary-ingress"),
			Namespace:       namespace.Metadata.Name().Elem(),
			CreateNamespace: pulumi.BoolPtr(false),
			RepositoryOpts: &helm.RepositoryOptsArgs{
				Repo: pulumi.String("https://app.getambassador.io"),
			},
			ForceUpdate: pulumi.Bool(true),
			Values: pulumi.Map{
				"replicaCount": pulumi.Int(1),
				"hostNetwork":  pulumi.Bool(false),
				"dnsPolicy":    pulumi.String("ClusterFirstWithHostNet"),
				"service": pulumi.Map{
					"type": pulumi.String("LoadBalancer"),
					"ports": pulumi.Array{
						pulumi.Map{
							"name":       pulumi.String("http"),
							"port":       pulumi.Int(80),
							"targetPort": pulumi.Int(8080),
						},
						pulumi.Map{
							"name":       pulumi.String("https"),
							"port":       pulumi.Int(443),
							"targetPort": pulumi.Int(8443),
						},
					},
					"annotations": pulumi.Map{
						"service.beta.kubernetes.io/aws-load-balancer-type":            pulumi.String("external"),
						"service.beta.kubernetes.io/aws-load-balancer-nlb-target-type": pulumi.String("instance"),
						"service.beta.kubernetes.io/aws-load-balancer-scheme":          pulumi.String("internet-facing"),
					},
				},
				"serviceAccount": pulumi.Map{
					"create": pulumi.Bool(false),
					"name":   sa.Metadata.Name(),
				},
				"agent": pulumi.Map{
					"enabled": pulumi.Bool(false),
				},
			},
		}, pulumi.Provider(kubeProvider), pulumi.DependsOn([]pulumi.Resource{crds, sa, namespace}))
		if err != nil {
			return err
		}

		_, err = apiextensions.NewCustomResource(ctx, "ingress-http-listener", &apiextensions.CustomResourceArgs{
			ApiVersion: pulumi.String("getambassador.io/v3alpha1"),
			Kind:       pulumi.String("Listener"),
			Metadata: metav1.ObjectMetaArgs{
				Name:      pulumi.String("ingress-http-listener"),
				Namespace: namespace.Metadata.Name().Elem(),
			},
			OtherFields: kubernetes.UntypedArgs{
				"spec": pulumi.Map{
					"port":          pulumi.Int(8080),
					"protocol":      pulumi.String("HTTP"),
					"securityModel": pulumi.String("XFP"),
					"hostBinding": pulumi.Map{
						"namespace": pulumi.Map{
							"from": pulumi.String("ALL"),
						},
					},
				},
			},
		}, pulumi.Provider(kubeProvider), pulumi.DependsOn([]pulumi.Resource{crds, namespace}))
		if err != nil {
			return err
		}

		_, err = apiextensions.NewCustomResource(ctx, "ingress-https-listener", &apiextensions.CustomResourceArgs{
			ApiVersion: pulumi.String("getambassador.io/v3alpha1"),
			Kind:       pulumi.String("Listener"),
			Metadata: metav1.ObjectMetaArgs{
				Name:      pulumi.String("ingress-https-listener"),
				Namespace: namespace.Metadata.Name().Elem(),
			},
			OtherFields: kubernetes.UntypedArgs{
				"spec": pulumi.Map{
					"port":          pulumi.Int(8443),
					"protocol":      pulumi.String("HTTPS"),
					"securityModel": pulumi.String("XFP"),
					"hostBinding": pulumi.Map{
						"namespace": pulumi.Map{
							"from": pulumi.String("ALL"),
						},
					},
				},
			},
		}, pulumi.Provider(kubeProvider), pulumi.DependsOn([]pulumi.Resource{crds, namespace}))
		if err != nil {
			return err
		}

		return nil
	})
}
