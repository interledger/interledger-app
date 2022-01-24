package ingress

import (
	"github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes"
	"github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes/apiextensions"
	v1 "github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes/meta/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type DeployHostArgs struct {
	Hostname  string
	TlsSecret string
}

func DeployHost(ctx *pulumi.Context, args *DeployHostArgs, opts ...pulumi.ResourceOption) error {

	spec := pulumi.Map{
		"hostname": pulumi.String("*"),
		"requestPolicy": pulumi.Map{
			"insecure": pulumi.Map{
				"action": pulumi.String("Route"),
			},
		},
	}

	if args.Hostname != "" {
		spec["hostname"] = pulumi.String(args.Hostname)
	}

	if args.TlsSecret != "" {
		spec["tlsSecret"] = pulumi.Map{
			"name": pulumi.String(args.TlsSecret),
		}
	}

	_, err := apiextensions.NewCustomResource(ctx, "ingress-host", &apiextensions.CustomResourceArgs{
		ApiVersion: pulumi.String("getambassador.io/v3alpha1"),
		Kind:       pulumi.String("Host"),
		Metadata: v1.ObjectMetaArgs{
			Name: pulumi.String("host"),
		},
		OtherFields: kubernetes.UntypedArgs{
			"spec": spec,
		},
	}, opts...)

	if err != nil {
		return err
	}

	return nil
}
