package cockroach

import (
	"github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes"
	"github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes/apiextensions"
	v1 "github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes/meta/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func createNodeCert(ctx *pulumi.Context, issuer string, ns string, serviceName string) (*apiextensions.CustomResource, error) {
	cr, err := apiextensions.NewCustomResource(ctx, "cockroach-node", &apiextensions.CustomResourceArgs{
		ApiVersion: pulumi.String("cert-manager.io/v1"),
		Kind:       pulumi.String("Certificate"),
		Metadata: v1.ObjectMetaArgs{
			Name: pulumi.String("cockroach-node"),
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
						pulumi.String("Cockroach"),
					},
				},
				"dnsNames": pulumi.StringArray{
					pulumi.String("localhost"),
					pulumi.String("127.0.0.1"),
					pulumi.Sprintf("%s-public", serviceName),
					pulumi.Sprintf("%s-public.%s", serviceName, ns),
					pulumi.Sprintf("%s-public.%s.svc.cluster.local", serviceName, ns),
					pulumi.Sprintf("*.%s", serviceName),
					pulumi.Sprintf("*.%s.%s", serviceName, ns),
					pulumi.Sprintf("*.%s.%s.svc.cluster.local", serviceName, ns),
				},
				"secretName": pulumi.String("cockroachdb-node"),
				"issuerRef": pulumi.Map{
					"name":  pulumi.String(issuer),
					"kind":  pulumi.String("Issuer"),
					"group": pulumi.String("cert-manager.io"),
				},
			},
		},
	})

	if err != nil {
		return nil, err
	}

	return cr, nil
}
