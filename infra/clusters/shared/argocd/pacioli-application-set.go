package main

import (
	"github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes"
	"github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes/apiextensions"
	metav1 "github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes/meta/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type pacioliApplicationSetArgs struct {
	Namespace pulumi.StringInput
}

func newPacioliApplicationSet(ctx *pulumi.Context, args pacioliApplicationSetArgs, opts ...pulumi.ResourceOption) error {

	_, err := apiextensions.NewCustomResource(ctx, "pacioli-application-set", &apiextensions.CustomResourceArgs{
		ApiVersion: pulumi.String("argoproj.io/v1alpha1"),
		Kind:       pulumi.String("ApplicationSet"),
		Metadata: metav1.ObjectMetaArgs{
			Name:      pulumi.String("pacioli"),
			Namespace: args.Namespace,
		},
		OtherFields: kubernetes.UntypedArgs{
			"spec": pulumi.Map{
				"generators": pulumi.MapArray{
					pulumi.Map{
						"list": pulumi.Map{
							"elements": pulumi.MapArray{
								pulumi.Map{
									"cluster": pulumi.String("dev-eu1"),
									"url":     pulumi.String("https://AF83FFCC8A31D16D2E1C9C1788205863.sk1.eu-west-1.eks.amazonaws.com"),
								},
							},
						},
					},
				},
				"template": pulumi.Map{
					"metadata": pulumi.Map{
						"name": pulumi.String("{{cluster}}-pacioli"),
					},
					"spec": pulumi.Map{
						"project": pulumi.String("default"),
						"source": pulumi.Map{
							"repoURL":        pulumi.String("https://gitlab.com/fynbos/rooibos.git"),
							"targetRevision": pulumi.String("main"),
							"path":           pulumi.String("pacioli/envs/{{cluster}}"),
						},
						"destination": pulumi.Map{
							"server":    pulumi.String("{{url}}"),
							"namespace": pulumi.String("pacioli"),
						},
						"syncPolicy": pulumi.Map{
							"automated": pulumi.Map{},
						},
					},
				},
			},
		},
	}, opts...)

	if err != nil {
		return err
	}

	return nil
}
