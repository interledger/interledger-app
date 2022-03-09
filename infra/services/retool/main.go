package retool

import (
	"github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes/helm/v3"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"gitlab.com/fynbos/infra/services/ingress"
)

func DeployRetool(ctx *pulumi.Context, hn string) (*helm.Chart, error) {
	values := pulumi.Map{
		"config": pulumi.Map{
			"licenseKey":         pulumi.String("SSOP_0645bcae-fd52-4ce5-ab33-8001d81b7d38"),
			"useInsecureCookies": pulumi.Bool(true),
			"encryptionKey":      pulumi.String("9i/IwIhNuuXczH9q6mU8YyqzVPY3eM3r7H3qBSdKJk9XHJkSLj1C3Zeey24zfMbdzlEzWsieZ5G+4vtJkK/F4w=="),
			"jwtSecret":          pulumi.String("9i/IwIhNuuXczH9q6mU8YyqzVPY3eM3r7H3qBSdKJk9XHJkSLj1C3Zeey24zfMbdzlEzWsieZ5G+4vtJkK/F4w=="),
		},
		"image": pulumi.Map{
			"tag": pulumi.String("latest"),
		},
	}

	chart, err := helm.NewChart(ctx, "retool", helm.ChartArgs{
		Version: pulumi.String("4.8.0"),
		Chart:   pulumi.String("retool"),
		FetchArgs: &helm.FetchArgs{
			Repo: pulumi.String("https://charts.retool.com"),
		},
		Values: values,
	})
	if err != nil {
		return nil, err
	}

	err = deployIngress(ctx, hn)
	if err != nil {
		return nil, err
	}

	return chart, nil
}

func deployIngress(ctx *pulumi.Context, hn string) error {
	err := ingress.DeployMapping(ctx, &ingress.MappingArgs{
		Name:            "retool",
		Hostname:        hn,
		Prefix:          "/",
		Rewrite:         "/",
		Service:         "retool:3000",
		EnableWebsocket: true,
	})
	if err != nil {
		return err
	}

	return nil
}
