package main

import (
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	cert_manager "gitlab.com/fynbos/infra/services/cert-manager"
	"gitlab.com/fynbos/infra/services/ingress"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {

		//err := cert_manager.DeployCertManager(ctx)
		//if err != nil {
		//	return err
		//}

		err := cert_manager.BootstrapCA(ctx)
		if err != nil {
			return err
		}

		err = ingress.DeployEmissaryIngress(ctx)
		if err != nil {
			return err
		}

		err = ingress.DeployHost(ctx)
		if err != nil {
			return err
		}
		err = ingress.DeployListeners(ctx)
		if err != nil {
			return err
		}

		//err = cockroach.DeployCockroach(ctx)
		//if err != nil {
		//	return err
		//}

		return nil
	})
}
