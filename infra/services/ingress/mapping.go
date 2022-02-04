package ingress

import (
	"fmt"
	"github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes"
	"github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes/apiextensions"
	v1 "github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes/meta/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type MappingArgs struct {
	Name            string
	Hostname        string
	Prefix          string
	Rewrite         string
	Service         string
	EnableWebsocket bool
}

func DeployMapping(ctx *pulumi.Context, args *MappingArgs, opts ...pulumi.ResourceOption) error {

	spec := pulumi.Map{
		"hostname": pulumi.String(args.Hostname),
		"prefix":   pulumi.String(args.Prefix),
		"rewrite":  pulumi.String(args.Rewrite),
		"service":  pulumi.String(args.Service),
	}

	if args.EnableWebsocket {
		spec["allow_upgrade"] = pulumi.StringArray{
			pulumi.String("websocket"),
		}
	}

	_, err := apiextensions.NewCustomResource(ctx, fmt.Sprintf("mapping-%s", args.Name), &apiextensions.CustomResourceArgs{
		ApiVersion: pulumi.String("getambassador.io/v3alpha1"),
		Kind:       pulumi.String("Mapping"),
		Metadata: v1.ObjectMetaArgs{
			Name: pulumi.String(args.Name),
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
