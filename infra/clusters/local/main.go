package main

import (
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	cert_manager "gitlab.com/fynbos/infra/services/cert-manager"
	"gitlab.com/fynbos/infra/services/cockroach"
	"gitlab.com/fynbos/infra/services/ingress"
	"gitlab.com/fynbos/infra/services/kratos"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {

		cmChart, err := cert_manager.DeployCertManager(ctx)
		if err != nil {
			return err
		}

		caResource, err := cert_manager.BootstrapCA(ctx, pulumi.DependsOnInputs(cmChart.Ready))
		if err != nil {
			return err
		}

		ingressChart, err := ingress.DeployEmissaryIngress(ctx)
		if err != nil {
			return err
		}

		// Depends on here is a workaround due to gremlins https://github.com/pulumi/pulumi-kubernetes/issues/861
		err = ingress.DeployHost(ctx, pulumi.DependsOnInputs(ingressChart.Ready))
		if err != nil {
			return err
		}
		err = ingress.DeployListeners(ctx, pulumi.DependsOnInputs(ingressChart.Ready))
		if err != nil {
			return err
		}

		err = cockroach.DeployCockroach(ctx, pulumi.DependsOn([]pulumi.Resource{caResource}))
		if err != nil {
			return err
		}

		_, err = kratos.DeployKratos(ctx)

		return nil
	})
}
