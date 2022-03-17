package main

import (
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"gitlab.com/fynbos/infra/services/backend"
	cert_manager "gitlab.com/fynbos/infra/services/cert-manager"
	"gitlab.com/fynbos/infra/services/cockroach"
	"gitlab.com/fynbos/infra/services/ingress"
	"gitlab.com/fynbos/infra/services/kratos"
	"gitlab.com/fynbos/infra/services/mailhog"
	pacioli "gitlab.com/fynbos/infra/services/pacioli"
	"gitlab.com/fynbos/infra/services/protea"
	"gitlab.com/fynbos/infra/services/retool"
	"gitlab.com/fynbos/infra/services/temporal"
	"gitlab.com/fynbos/infra/services/tigerbeetle"
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
			Name:     "ingress",
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
		_, err = kratos.DeployKratos(ctx, crCert, "http://fynbos.test")
		if err != nil {
			return err
		}
		err = kratos.DeployKratosIngress(ctx, pulumi.DependsOnInputs(ingressChart.Ready))

		err = mailhog.DeployMailHog(ctx, "mail.fynbos.test")
		if err != nil {
			return err
		}

		err = temporal.DeployTemporalDev(ctx, "localhost:5005", "latest")
		if err != nil {
			return err
		}

		err = protea.DeployProtea(ctx, protea.DeployProteaArgs{
			ImageRepo: "localhost:5005",
			ImageTag:  "latest",
		})
		if err != nil {
			return err
		}

		beCert, err := cockroach.CreateClientCert(ctx, &cockroach.ClientCertArgs{
			Issuer:    "ca-issuer",
			Namespace: "default",
			Name:      "backend",
		}, pulumi.DependsOn([]pulumi.Resource{caResource}))
		if err != nil {
			return err
		}

		err = backend.DeployBackend(ctx, backend.DeployBackendArgs{
			ImageRepo:            "localhost:5005",
			Cert:                 beCert,
			ImageTag:             "latest",
			UsdLedgerCode:        0,
			NoopEquityAccountID:  "036c9b47-d0e4-4960-863e-a80224aa6ff3",
			EnablePlayground:     true,
			Hostname:             "fynbos.test",
			GoogleOauth2ClientID: "572950914705-dv7oqq4r8bqljv3s831qqcan1n6f8vvs.apps.googleusercontent.com",
		})
		if err != nil {
			return err
		}

		err = tigerbeetle.DeployTigerBeetle(ctx, tigerbeetle.DeployTigerBeetleArgs{
			IsLocal: true,
		})
		if err != nil {
			return err
		}

		pcCert, err := cockroach.CreateClientCert(ctx, &cockroach.ClientCertArgs{
			Name:      "pacioli",
			Issuer:    "ca-issuer",
			Namespace: "default",
		}, pulumi.DependsOn([]pulumi.Resource{caResource}))
		if err != nil {
			return err
		}
		err = pacioli.DeployPacioli(ctx, &pacioli.DeployPacioliArgs{
			Cert:      pcCert,
			ImageRepo: "localhost:5005",
			ImageTag:  "latest",
			Namespace: "default",
		})
		if err != nil {
			return err
		}

		err = ingress.DeployHost(ctx, &ingress.DeployHostArgs{
			Name:     "retool-ingress",
			Hostname: "localhost", // use localhost so Google will redirect.
		}, pulumi.DependsOnInputs(ingressChart.Ready))
		_, err = retool.DeployRetool(ctx, "localhost")
		if err != nil {
			return err
		}

		return nil
	})
}
