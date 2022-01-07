package main

import (
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	cert_manager "gitlab.com/fynbos/infra/services/cert-manager"
	"gitlab.com/fynbos/infra/services/cockroach"
	"gitlab.com/fynbos/infra/services/ingress"
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

		ingressChart, err := ingress.DeployEmissaryIngress(ctx, ingress.EmissaryIngressArgs{
			ReplicaCount: 1,
			Service: pulumi.Map{
				"type": pulumi.String("LoadBalancer"),
				"ports": pulumi.Array{
					pulumi.Map{
						"name":       pulumi.String("http"),
						"port":       pulumi.Int(80),
						"targetPort": pulumi.Int(8080),
					},
					pulumi.Map{
						"name":       pulumi.String("https"),
						"port":       pulumi.Int(443),
						"targetPort": pulumi.Int(8443),
					},
				},
				"annotations": pulumi.Map{
					"service.beta.kubernetes.io/aws-load-balancer-type":                              pulumi.String("nlb"),
					"service.beta.kubernetes.io/aws-load-balancer-cross-zone-load-balancing-enabled": pulumi.String("true"),
				},
			},
		})
		if err != nil {
			return err
		}

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

		return nil
	})
}
