package cert_manager

import (
	"github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes/helm/v3"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func DeployCertManager(ctx *pulumi.Context) (*helm.Chart, error) {

	chart, err := helm.NewChart(ctx, "cert-manager", helm.ChartArgs{
		Version: pulumi.String("1.6.1"),
		Chart:   pulumi.String("cert-manager"),
		FetchArgs: &helm.FetchArgs{
			Repo: pulumi.String("https://charts.jetstack.io"),
		},
		Values: pulumi.Map{
			"installCRDs": pulumi.Bool(true),
		},
	})

	if err != nil {
		return nil, err
	}

	return chart, nil
}
