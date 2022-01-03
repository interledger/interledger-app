package ingress

import (
	"github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes"
	"github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes/apiextensions"
	v1 "github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes/meta/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func DeployListeners(ctx *pulumi.Context, opts ...pulumi.ResourceOption) error {

	_, err := apiextensions.NewCustomResource(ctx, "ingress-http-listener", &apiextensions.CustomResourceArgs{
		ApiVersion: pulumi.String("getambassador.io/v3alpha1"),
		Kind:       pulumi.String("Listener"),
		Metadata: v1.ObjectMetaArgs{
			Name: pulumi.String("ingress-http-listener"),
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
	}, opts...)
	if err != nil {
		return err
	}

	_, err = apiextensions.NewCustomResource(ctx, "ingress-https-listener", &apiextensions.CustomResourceArgs{
		ApiVersion: pulumi.String("getambassador.io/v3alpha1"),
		Kind:       pulumi.String("Listener"),
		Metadata: v1.ObjectMetaArgs{
			Name: pulumi.String("ingress-https-listener"),
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
	}, opts...)
	if err != nil {
		return err
	}

	return nil
}
