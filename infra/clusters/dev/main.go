package main

import (
	"encoding/base64"
	b64 "encoding/base64"
	"fmt"
	"github.com/pulumi/pulumi-random/sdk/v4/go/random"
	"gitlab.com/fynbos/infra/services/postgres"
	"gitlab.com/fynbos/infra/services/rafiki"
	"os/exec"
	"strings"

	"os"

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
	"gitlab.com/fynbos/infra/services/redis"
	"gitlab.com/fynbos/infra/services/retool"
	"gitlab.com/fynbos/infra/services/temporal"
	"gitlab.com/fynbos/infra/services/tigerbeetle"
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
		_, err = kratos.DeployKratos(ctx, crCert, "https://dev.fynbos.dev", "CHANGE-ME-I-AM-VERY-INSECURE1234")
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

		err = temporal.DeployTemporalDev(ctx, ecrRepo, "latest")
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

		beCert, err := cockroach.CreateClientCert(ctx, &cockroach.ClientCertArgs{
			Issuer:    "ca-issuer",
			Namespace: "default",
			Name:      "backend",
		}, pulumi.DependsOn([]pulumi.Resource{caResource}))
		if err != nil {
			return err
		}
		err = backend.DeployBackend(ctx, backend.DeployBackendArgs{
			ImageRepo:            ecrRepo,
			Cert:                 beCert,
			ImageTag:             hash,
			UsdLedgerCode:        100,
			NoopEquityAccountID:  "7c63db4d-2f4c-4ab2-935e-7482bad12649",
			UnitToken:            "todo token",
			EnablePlayground:     true,
			Hostname:             "dev.fynbos.dev",
			GoogleOauth2ClientID: "572950914705-ith2keqq6l3cu652n262jd0gf9ffi7ka.apps.googleusercontent.com",
		})
		if err != nil {
			return err
		}

		err = tigerbeetle.DeployTigerBeetle(ctx, tigerbeetle.DeployTigerBeetleArgs{
			IsLocal:   false,
			ImageRepo: ecrRepo,
			ImageTag:  "patch-1",
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
			ImageRepo: ecrRepo,
			ImageTag:  hash,
			Namespace: "default",
		})
		if err != nil {
			return err
		}

		err = ingress.DeployHost(ctx, &ingress.DeployHostArgs{
			Name:      "retool-ingress",
			Hostname:  "retool.fynbos.dev",
			TlsSecret: "tls-secret",
		}, pulumi.DependsOnInputs(ingressChart.Ready))
		if err != nil {
			return err
		}
		_, err = retool.DeployRetool(ctx, retool.Args{
			Hostname:       "retool.fynbos.dev",
			CreatePostgres: true,
		})
		if err != nil {
			return err
		}

		err = redis.DeployRedis(ctx, &redis.DeployRedisArgs{
			EnableSentinal: false,
			ReplicaCount:   0,
		}, pulumi.DependsOn([]pulumi.Resource{caResource}))
		if err != nil {
			return err
		}

		postgresUserPassword, err := random.NewRandomPassword(ctx, "rafiki-postgres", &random.RandomPasswordArgs{
			Length:  pulumi.Int(32),
			Special: pulumi.Bool(true),
		})
		if err != nil {
			return err
		}
		err = postgres.DeployPostgres(ctx, &postgres.DeployPostgresArgs{
			ReadReplicasCount:    0,
			PostgresUserPassword: postgresUserPassword,
		}, pulumi.DependsOn([]pulumi.Resource{caResource}))
		if err != nil {
			return err
		}

		err = ingress.DeployHost(ctx, &ingress.DeployHostArgs{
			Name:      "rafiki-ingress",
			Hostname:  "pay.fynbos.dev",
			TlsSecret: "tls-secret",
		}, pulumi.DependsOnInputs(ingressChart.Ready))
		if err != nil {
			return err
		}

		streamSecret, err := random.NewRandomString(ctx, "rafiki-stream-secret", &random.RandomStringArgs{
			Length: pulumi.Int(32),
		}, pulumi.AdditionalSecretOutputs([]string{"Result"}))
		if err != nil {
			return err
		}
		streamSecretBase64 := streamSecret.Result.ApplyT(func(str string) string {
			return base64.RawStdEncoding.EncodeToString([]byte(str)[0:32])
		}).(pulumi.StringInput)

		rafikiPostgresCert, err := postgres.CreateClientCert(ctx, &postgres.ClientCertArgs{
			Name:      "rafiki",
			Issuer:    "ca-issuer",
			Namespace: "default",
		}, pulumi.DependsOn([]pulumi.Resource{caResource}))
		if err != nil {
			return err
		}
		rafikiRedisCert, err := redis.CreateClientCert(ctx, &redis.ClientCertArgs{
			Name:      "rafiki",
			Issuer:    "ca-issuer",
			Namespace: "default",
		}, pulumi.DependsOn([]pulumi.Resource{caResource}))
		if err != nil {
			return err
		}
		err = rafiki.DeployRafiki(ctx, &rafiki.DeployRafikiArgs{
			Name:                    "rafiki",
			DeployPlaygroundIngress: false,
			ImageRepo:               fmt.Sprintf("%s/rafiki-backend", ecrRepo),
			ImageTag:                "latest",
			DbBaseUrl:               "postgresql://rafiki@postgres-postgresql/rafiki",
			DbCert:                  rafikiPostgresCert,
			TbClusterID:             "0",
			// TbReplicaAddresses:         "[\"tigerbeetle-0.tigerbeetle.default.svc.cluster.local\"]",
			// waiting for update for client to handle dns.
			// you will need to manually look up the tigerbeetle-0 pod and get its ip address for now.
			TbReplicaAddresses:         "[\"10.100.71.147:8080\"]",
			IlpAddress:                 "test.fynbos",
			NonceRedisKey:              "noncefynbos",
			RedisUrl:                   "redis://redis-master:6379",
			RedisCert:                  rafikiRedisCert,
			AuthServerGrantUrl:         "http://127.0.0.1:3006",
			AuthServerIntrospectionUrl: "http://127.0.0.1:3007",
			StreamSecret:               streamSecretBase64,
			AdminKey:                   streamSecret.Result, // re-using for local cluster
			Hostname:                   "pay.fynbos.dev",
			PublicHost:                 "https://pay.fynbos.dev",
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
