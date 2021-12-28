package ingress

import (
	"github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes/helm/v3"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func DeployEmissaryIngress(ctx *pulumi.Context) error {

	_, err := helm.NewChart(ctx, "emissary", helm.ChartArgs{
		Version: pulumi.String("7.1.9"),
		Chart:   pulumi.String("emissary-ingress"),
		FetchArgs: &helm.FetchArgs{
			Repo: pulumi.String("https://app.getambassador.io"),
		},
		Values: pulumi.Map{
			"replicaCount": pulumi.Int(1),
			"service": pulumi.Map{
				"type": pulumi.String("NodePort"),
				"ports": pulumi.Array{
					pulumi.Map{
						"name":       pulumi.String("http"),
						"port":       pulumi.Int(8080),
						"hostPort":   pulumi.Int(8080),
						"targetPort": pulumi.Int(8080),
					},
					pulumi.Map{
						"name":       pulumi.String("https"),
						"port":       pulumi.Int(8443),
						"hostPort":   pulumi.Int(8443),
						"targetPort": pulumi.Int(8443),
					},
				},
			},
		},
	})

	if err != nil {
		return err
	}

	return nil
}
