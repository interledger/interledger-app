package main

import (
	b64 "encoding/base64"
	"errors"
	"os/exec"
	"strings"

	v1 "github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes/core/v1"
	metav1 "github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes/meta/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"
	"gitlab.com/fynbos/infra/services/backend"
	cert_manager "gitlab.com/fynbos/infra/services/cert-manager"
	"gitlab.com/fynbos/infra/services/cockroach"
	"gitlab.com/fynbos/infra/services/ingress"
	"gitlab.com/fynbos/infra/services/kratos"
	"gitlab.com/fynbos/infra/services/mailhog"
	"gitlab.com/fynbos/infra/services/pacioli"
	"gitlab.com/fynbos/infra/services/protea"
	"gitlab.com/fynbos/infra/services/tigerbeetle"
	"os"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		hash, err := getShortHash()
		if err != nil {
			return err
		}
		cfg := config.New(ctx, "fynbos")
		ecrRepo := cfg.Get("ecrRepo")
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

		cf, err := pulumi.NewStackReference(ctx, "fynbos/cf-fynbos.dev/main", nil)
		if err != nil {
			return err
		}

		pulumi.All(cf.GetStringOutput(pulumi.String("devClusterCert")), cf.GetStringOutput(pulumi.String("devClusterPrivateKey"))).ApplyT(
			func(args []interface{}) (*v1.Secret, error) {
				cert := args[0].(string)
				key := args[1].(string)

				b64Cert := b64.StdEncoding.EncodeToString([]byte(cert))
				b64Key := b64.StdEncoding.EncodeToString([]byte(key))
				return v1.NewSecret(ctx, "tls-secret", &v1.SecretArgs{
					Metadata: metav1.ObjectMetaArgs{
						Name: pulumi.String("tls-secret"),
					},
					Type: pulumi.String("kubernetes.io/tls"),
					Data: pulumi.StringMap{
						"tls.crt": pulumi.String(b64Cert),
						"tls.key": pulumi.String(b64Key),
					},
				})
			},
		)

		err = ingress.DeployHost(ctx, &ingress.DeployHostArgs{
			Name:      "ingress",
			Hostname:  "dev.fynbos.dev",
			TlsSecret: "tls-secret",
		}, pulumi.DependsOnInputs(ingressChart.Ready))
		err = ingress.DeployHost(ctx, &ingress.DeployHostArgs{
			Name:      "mail",
			Hostname:  "mail.fynbos.dev",
			TlsSecret: "tls-secret",
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
		if err != nil {
			return err
		}
		_, err = kratos.DeployKratos(ctx, crCert, "https://dev.fynbos.dev")
		if err != nil {
			return err
		}
		err = kratos.DeployKratosIngress(ctx, pulumi.DependsOnInputs(ingressChart.Ready))
		if err != nil {
			return err
		}

		err = mailhog.DeployMailHog(ctx, "mail.fynbos.dev")
		if err != nil {
			return err
		}

		err = protea.DeployProtea(ctx, protea.DeployProteaArgs{
			ImageRepo: ecrRepo,
			ImageTag:  hash,
		})
		if err != nil {
			return err
		}

		ledgerCodes := pacioli.DevClusterLedgerCodes()
		backendLedgerCode, present := ledgerCodes["backend-usd"]
		if !present {
			return errors.New("Ledger code for backend-usd does not exist.")
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
			ImageRepo:        ecrRepo,
			Cert:             beCert,
			ImageTag:         hash,
			UsdLedgerCode:    backendLedgerCode,
			EnablePlayground: true,
			Hostname:         "dev.fynbos.dev",
		})
		if err != nil {
			return err
		}

		err = tigerbeetle.DeployTigerBeetle(ctx, tigerbeetle.DeployTigerBeetleArgs{
			IsLocal: false,
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
			Cert:              pcCert,
			ImageRepo:         ecrRepo,
			ImageTag:          hash,
			Namespace:         "default",
			BackendLedgerCode: backendLedgerCode,
		})
		if err != nil {
			return err
		}

		return nil
	})
}

func getShortHash() (string, error) {
	envHash, present := os.LookupEnv("GIT_HASH")
	if present {
		return envHash, nil
	}

	out, err := exec.Command("git", "log", "-1", "--pretty=%h").Output()
	if err != nil {
		return "", err
	}

	return strings.TrimSuffix(string(out), "\n"), nil
}
