package main

import (
	"github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes"
	"github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes/apiextensions"
	metav1 "github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes/meta/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func NewDevProject(ctx *pulumi.Context, namespace pulumi.StringPtrInput, opts ...pulumi.ResourceOption) error {
	_, err := apiextensions.NewCustomResource(ctx, "dev-project", &apiextensions.CustomResourceArgs{
		ApiVersion: pulumi.String("argoproj.io/v1alpha1"),
		Kind:       pulumi.String("AppProject"),
		Metadata: metav1.ObjectMetaArgs{
			Name:      pulumi.String("dev"),
			Namespace: namespace,
			Finalizers: pulumi.StringArray{
				pulumi.String("resources-finalizer.argocd.argoproj.io"),
			},
		},
		OtherFields: kubernetes.UntypedArgs{
			"spec": pulumi.Map{
				"description": pulumi.String("Dev Project"),
				"sourceRepos": pulumi.StringArray{
					pulumi.String("https://gitlab.com/fynbos/rooibos.git"),
				},
				"destinations": pulumi.MapArray{
					pulumi.Map{
						"namespace": pulumi.String("*"),
						"server":    pulumi.String("https://AF83FFCC8A31D16D2E1C9C1788205863.sk1.eu-west-1.eks.amazonaws.com"),
					},
				},
				"clusterResourceWhitelist": pulumi.MapArray{
					pulumi.Map{
						"group": pulumi.String(""),
						"kind":  pulumi.String("Namespace"),
					},
				},
			},
		},
	}, opts...)

	return err
}
