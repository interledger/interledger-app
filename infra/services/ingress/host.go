package ingress

import (
	"github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes"
	"github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes/apiextensions"
	v1 "github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes/meta/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func DeployHost(ctx *pulumi.Context, opts ...pulumi.ResourceOption) error {

	_, err := apiextensions.NewCustomResource(ctx, "ingress-host", &apiextensions.CustomResourceArgs{
		ApiVersion: pulumi.String("getambassador.io/v3alpha1"),
		Kind:       pulumi.String("Host"),
		Metadata: v1.ObjectMetaArgs{
			Name: pulumi.String("host"),
		},
		OtherFields: kubernetes.UntypedArgs{
			"spec": pulumi.Map{
				"hostname": pulumi.String("*"),
				"requestPolicy": pulumi.Map{
					"insecure": pulumi.Map{
						"action": pulumi.String("Route"),
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
