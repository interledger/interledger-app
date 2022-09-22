package main

import (
	"github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes"
	"github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes/apiextensions"
	metav1 "github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes/meta/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type postgresApplicationSetArgs struct {
	Namespace pulumi.StringInput
}

func newPostgresApplicationSet(ctx *pulumi.Context, args postgresApplicationSetArgs, opts ...pulumi.ResourceOption) (*apiextensions.CustomResource, error) {

	postgresApplicationSet, err := apiextensions.NewCustomResource(ctx, "postgres-application-set", &apiextensions.CustomResourceArgs{
		ApiVersion: pulumi.String("argoproj.io/v1alpha1"),
		Kind:       pulumi.String("ApplicationSet"),
		Metadata: metav1.ObjectMetaArgs{
			Name:      pulumi.String("postgres"),
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
									"project": pulumi.String("dev"),
								},
							},
						},
					},
				},
				"template": pulumi.Map{
					"metadata": pulumi.Map{
						"name": pulumi.String("{{cluster}}-postgres"),
					},
					"spec": pulumi.Map{
						"project": pulumi.String("{{project}}"),
						"source": pulumi.Map{
							"repoURL":        pulumi.String("https://gitlab.com/fynbos/rooibos.git"),
							"targetRevision": pulumi.String("main"),
							"path":           pulumi.String("postgres/envs/{{cluster}}"),
						},
						"destination": pulumi.Map{
							"server":    pulumi.String("{{url}}"),
							"namespace": pulumi.String("postgres"),
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
		return nil, err
	}

	return postgresApplicationSet, nil
}
