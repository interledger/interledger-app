package ingress

import (
	"github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes/helm/v3"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type EmissaryIngressArgs struct {
	ReplicaCount int
	Service      pulumi.Map
}

func DeployEmissaryIngress(ctx *pulumi.Context, args EmissaryIngressArgs) (*helm.Chart, error) {
	chart, err := helm.NewChart(ctx, "emissary", helm.ChartArgs{
		Version: pulumi.String("7.1.9"),
		Chart:   pulumi.String("emissary-ingress"),
		FetchArgs: &helm.FetchArgs{
			Repo: pulumi.String("https://app.getambassador.io"),
		},
		Values: pulumi.Map{
			"replicaCount": pulumi.Int(args.ReplicaCount),
			"service":      args.Service,
		},
	})

	if err != nil {
		return nil, err
	}

	return chart, nil
}
