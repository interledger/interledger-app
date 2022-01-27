package cert_manager

import (
	"github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes"
	"github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes/apiextensions"
	v1 "github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes/meta/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func BootstrapCA(ctx *pulumi.Context, opts ...pulumi.ResourceOption) (*apiextensions.CustomResource, error) {
	_, err := createClusterIssuer(ctx, opts...)
	if err != nil {
		return nil, err
	}

	err = createCert(ctx, opts...)
	if err != nil {
		return nil, err
	}

	cr, err := createIssuer(ctx, opts...)
	if err != nil {
		return nil, err
	}

	return cr, nil
}

func createClusterIssuer(ctx *pulumi.Context, opts ...pulumi.ResourceOption) (*apiextensions.CustomResource, error) {
	cr, err := apiextensions.NewCustomResource(ctx, "cluster-issuer", &apiextensions.CustomResourceArgs{
		ApiVersion: pulumi.String("cert-manager.io/v1"),
		Kind:       pulumi.String("ClusterIssuer"),
		Metadata: v1.ObjectMetaArgs{
			Name: pulumi.String("selfsigned-issuer"),
		},
		OtherFields: kubernetes.UntypedArgs{
			"spec": pulumi.Map{
				"selfSigned": pulumi.Map{},
			},
		},
	}, opts...)

	if err != nil {
		return nil, err
	}

	return cr, nil
}

func createCert(ctx *pulumi.Context, opts ...pulumi.ResourceOption) error {
	_, err := apiextensions.NewCustomResource(ctx, "selfsigned-ca", &apiextensions.CustomResourceArgs{
		ApiVersion: pulumi.String("cert-manager.io/v1"),
		Kind:       pulumi.String("Certificate"),
		Metadata: v1.ObjectMetaArgs{
			Name: pulumi.String("selfsigned-ca"),
		},
		OtherFields: kubernetes.UntypedArgs{
			"spec": pulumi.Map{
				"isCA":       pulumi.Bool(true),
				"commonName": pulumi.String("selfsigned-ca"),
				"secretName": pulumi.String("root-secret"),
				"privateKey": pulumi.Map{
					"algorithm": pulumi.String("ECDSA"),
					"size":      pulumi.Int(256),
				},
				"issuerRef": pulumi.Map{
					"name":  pulumi.String("selfsigned-issuer"),
					"kind":  pulumi.String("ClusterIssuer"),
					"group": pulumi.String("cert-manager.io"),
				},
			},
		},
	}, opts...)

	if err != nil {
		return err
	}

	return nil
}

func createIssuer(ctx *pulumi.Context, opts ...pulumi.ResourceOption) (*apiextensions.CustomResource, error) {
	cr, err := apiextensions.NewCustomResource(ctx, "ca-issuer", &apiextensions.CustomResourceArgs{
		ApiVersion: pulumi.String("cert-manager.io/v1"),
		Kind:       pulumi.String("Issuer"),
		Metadata: v1.ObjectMetaArgs{
			Name: pulumi.String("ca-issuer"),
		},
		OtherFields: kubernetes.UntypedArgs{
			"spec": pulumi.Map{
				"ca": pulumi.Map{
					"secretName": pulumi.String("root-secret"),
				},
			},
		},
	}, opts...)
	if err != nil {
		return nil, err
	}

	return cr, nil
}
