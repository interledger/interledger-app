package redis

import (
	"fmt"

	"github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes"
	"github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes/apiextensions"
	v1 "github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes/meta/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type NodeCertArgs struct {
	Issuer      string
	Namespace   string
	ServiceName string
}

func createNodeCert(ctx *pulumi.Context, args *NodeCertArgs, opts ...pulumi.ResourceOption) (*apiextensions.CustomResource, error) {
	cr, err := apiextensions.NewCustomResource(ctx, "redis-node", &apiextensions.CustomResourceArgs{
		ApiVersion: pulumi.String("cert-manager.io/v1"),
		Kind:       pulumi.String("Certificate"),
		Metadata: v1.ObjectMetaArgs{
			Name: pulumi.String("redis-node"),
		},
		OtherFields: kubernetes.UntypedArgs{
			"spec": pulumi.Map{
				"duration":    pulumi.String("8760h"),
				"renewBefore": pulumi.String("168h"),
				"usages": pulumi.StringArray{
					pulumi.String("digital signature"),
					pulumi.String("key encipherment"),
					pulumi.String("server auth"),
					pulumi.String("client auth"),
				},
				"privateKey": pulumi.Map{
					"algorithm": pulumi.String("RSA"),
					"size":      pulumi.Int(2048),
				},
				"commonName": pulumi.String("node"),
				"subject": pulumi.Map{
					"organizations": pulumi.StringArray{
						pulumi.String("Redis"),
					},
				},
				"dnsNames": pulumi.StringArray{
					pulumi.String("localhost"),
					pulumi.String("127.0.0.1"),
					pulumi.Sprintf("%s", args.ServiceName),
					pulumi.Sprintf("%s.%s", args.ServiceName, args.Namespace),
					pulumi.Sprintf("%s.%s.svc.cluster.local", args.ServiceName, args.Namespace),
					pulumi.Sprintf("*.%s", args.ServiceName),
					pulumi.Sprintf("*.%s.%s", args.ServiceName, args.Namespace),
					pulumi.Sprintf("*.%s.%s.svc.cluster.local", args.ServiceName, args.Namespace),
				},
				"secretName": pulumi.String("redis-node"),
				"issuerRef": pulumi.Map{
					"name":  pulumi.String(args.Issuer),
					"kind":  pulumi.String("Issuer"),
					"group": pulumi.String("cert-manager.io"),
				},
			},
		},
	}, opts...)

	if err != nil {
		return nil, err
	}

	return cr, nil
}

type ClientCertArgs struct {
	Name      string
	Issuer    string
	Namespace string
}

func CreateClientCert(ctx *pulumi.Context, args *ClientCertArgs, opts ...pulumi.ResourceOption) (*apiextensions.CustomResource, error) {
	cr, err := apiextensions.NewCustomResource(ctx, fmt.Sprintf("redis-%s-client", args.Name), &apiextensions.CustomResourceArgs{
		ApiVersion: pulumi.String("cert-manager.io/v1"),
		Kind:       pulumi.String("Certificate"),
		Metadata: v1.ObjectMetaArgs{
			Name: pulumi.Sprintf("redis-%s-client", args.Name),
		},
		OtherFields: kubernetes.UntypedArgs{
			"spec": pulumi.Map{
				"duration":    pulumi.String("672h"),
				"renewBefore": pulumi.String("48h"),
				"usages": pulumi.StringArray{
					pulumi.String("digital signature"),
					pulumi.String("key encipherment"),
					pulumi.String("client auth"),
				},
				"privateKey": pulumi.Map{
					"algorithm": pulumi.String("RSA"),
					"size":      pulumi.Int(2048),
				},
				"commonName": pulumi.String(args.Name),
				"subject": pulumi.Map{
					"organizations": pulumi.StringArray{
						pulumi.String("Redis"),
					},
				},
				"secretName": pulumi.Sprintf("redis-%s", args.Name),
				"issuerRef": pulumi.Map{
					"name":  pulumi.String(args.Issuer),
					"kind":  pulumi.String("Issuer"),
					"group": pulumi.String("cert-manager.io"),
				},
			},
		},
	}, opts...)

	if err != nil {
		return nil, err
	}

	return cr, nil
}
