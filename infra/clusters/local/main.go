package main

import (
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	cert_manager "gitlab.com/fynbos/infra/services/cert-manager"
	"gitlab.com/fynbos/infra/services/cockroach"
	"gitlab.com/fynbos/infra/services/ingress"
	"gitlab.com/fynbos/infra/services/kratos"
	"gitlab.com/fynbos/infra/services/mailhog"
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
		})
		if err != nil {
			return err
		}

		// Depends on here is a workaround due to gremlins https://github.com/pulumi/pulumi-kubernetes/issues/861
		err = ingress.DeployHost(ctx, &ingress.DeployHostArgs{
			Hostname: "fynbos.test",
		}, pulumi.DependsOnInputs(ingressChart.Ready))
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

		crCert, err := cockroach.CreateClientCert(ctx, &cockroach.ClientCertArgs{
			Issuer:    "ca-issuer",
			Namespace: "default",
			Name:      "kratos",
		}, pulumi.DependsOn([]pulumi.Resource{caResource}))
		_, err = kratos.DeployKratos(ctx, crCert)
		if err != nil {
			return err
		}
		err = kratos.DeployKratosIngress(ctx, pulumi.DependsOnInputs(ingressChart.Ready))

		err = mailhog.DeployMailHog(ctx)
		if err != nil {
			return err
		}

		_, err = cockroach.CreateClientCert(ctx, &cockroach.ClientCertArgs{
			Issuer:    "ca-issuer",
			Namespace: "default",
			Name:      "backend",
		}, pulumi.DependsOn([]pulumi.Resource{caResource}))
		if err != nil {
			return err
		}

		return nil
	})
}
