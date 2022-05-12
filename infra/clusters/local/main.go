package main

import (
	"encoding/base64"

	"github.com/pulumi/pulumi-random/sdk/v4/go/random"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"gitlab.com/fynbos/infra/services/backend"
	cert_manager "gitlab.com/fynbos/infra/services/cert-manager"
	"gitlab.com/fynbos/infra/services/cockroach"
	"gitlab.com/fynbos/infra/services/ingress"
	"gitlab.com/fynbos/infra/services/kratos"
	"gitlab.com/fynbos/infra/services/mailhog"
	mockstaticresponseservers "gitlab.com/fynbos/infra/services/mock-static-response-servers"
	pacioli "gitlab.com/fynbos/infra/services/pacioli"
	"gitlab.com/fynbos/infra/services/postgres"
	"gitlab.com/fynbos/infra/services/protea"
	"gitlab.com/fynbos/infra/services/rafiki"
	"gitlab.com/fynbos/infra/services/redis"
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
		_, err = kratos.DeployKratos(ctx, crCert, "http://fynbos.test", "CHANGE-ME-I-AM-VERY-INSECURE1234")
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
			UnitToken:            "todo token",
			UnitBaseUrl:          "https://api.s.unit.sh",
			EnablePlayground:     true,
			Hostname:             "fynbos.test",
			GoogleOauth2ClientID: "572950914705-dv7oqq4r8bqljv3s831qqcan1n6f8vvs.apps.googleusercontent.com",
		})
		if err != nil {
			return err
		}

		err = tigerbeetle.DeployTigerBeetle(ctx, tigerbeetle.DeployTigerBeetleArgs{
			IsLocal:   true,
			ImageRepo: "localhost:5005",
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
		_, err = retool.DeployRetool(ctx, retool.Args{
			Hostname:       "losthost",
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
			Name:     "rafiki-ingress",
			Hostname: "pay.fynbos.test",
		}, pulumi.DependsOnInputs(ingressChart.Ready))
		if err != nil {
			return err
		}

		streamSecret, err := random.NewRandomString(ctx, "rafiki-stream-secret", &random.RandomStringArgs{
			Length: pulumi.Int(32),
		})
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
		})
		if err != nil {
			return err
		}
		rafikiRedisCert, err := redis.CreateClientCert(ctx, &redis.ClientCertArgs{
			Name:      "rafiki",
			Issuer:    "ca-issuer",
			Namespace: "default",
		})
		if err != nil {
			return err
		}

		// Fynbos' rafiki instance.
		err = rafiki.DeployRafiki(ctx, &rafiki.DeployRafikiArgs{
			Name:                    "rafiki",
			DeployPlaygroundIngress: true,
			ImageRepo:               "localhost:5005/rafiki-backend",
			ImageTag:                "latest",
			DbBaseUrl:               "postgresql://rafiki@postgres-postgresql/rafiki",
			DbCert:                  rafikiPostgresCert,
			TbClusterID:             "0",
			// TbReplicaAddresses:         "[\"tigerbeetle-0.tigerbeetle.default.svc.cluster.local\"]",
			// waiting for update for client to handle dns.
			// you will need to manually look up the tigerbeetle-0 pod and get its ip address for now.
			TbReplicaAddresses:         "[\"10.244.0.23:8080\"]",
			IlpAddress:                 "test.fynbos",
			NonceRedisKey:              "noncefynbos",
			RedisUrl:                   "redis://redis-master:6379/0",
			RedisCert:                  rafikiRedisCert,
			AuthServerGrantUrl:         "http://mockauth",
			AuthServerIntrospectionUrl: "http://mockauth",
			StreamSecret:               streamSecretBase64,
			AdminKey:                   streamSecret.Result,                     // re-using for local cluster
			WebhookUrl:                 "https://end3sf6r22xva.x.pipedream.net", // using request bin for now
			Hostname:                   "pay.fynbos.test",
			PublicHost:                 "http://rafiki", // using rafiki as coredns will find rafiki
		})
		if err != nil {
			return err
		}

		// Peer rafiki instance for dev testing.
		err = ingress.DeployHost(ctx, &ingress.DeployHostArgs{
			Name:     "peer-ingress",
			Hostname: "peer.fynbos.test",
		}, pulumi.DependsOnInputs(ingressChart.Ready))
		if err != nil {
			return err
		}
		peerPostgresCert, err := postgres.CreateClientCert(ctx, &postgres.ClientCertArgs{
			Name:      "peer",
			Issuer:    "ca-issuer",
			Namespace: "default",
		})
		if err != nil {
			return err
		}
		peerRedisCert, err := redis.CreateClientCert(ctx, &redis.ClientCertArgs{
			Name:      "peer",
			Issuer:    "ca-issuer",
			Namespace: "default",
		})
		if err != nil {
			return err
		}
		err = rafiki.DeployRafiki(ctx, &rafiki.DeployRafikiArgs{
			Name:                    "peer",
			DeployPlaygroundIngress: true,
			ImageRepo:               "localhost:5005/rafiki-backend",
			ImageTag:                "latest",
			DbBaseUrl:               "postgresql://peer@postgres-postgresql/peer",
			DbCert:                  peerPostgresCert,
			TbClusterID:             "0",
			// TbReplicaAddresses:         "[\"tigerbeetle-0.tigerbeetle.default.svc.cluster.local\"]",
			// waiting for update for client to handle dns.
			// you will need to manually look up the tigerbeetle-0 pod and get its ip address for now.
			TbReplicaAddresses:         "[\"10.244.0.23:8080\"]",
			IlpAddress:                 "test.peer",
			NonceRedisKey:              "noncepeer",
			RedisUrl:                   "redis://redis-master:6379/1",
			RedisCert:                  peerRedisCert,
			AuthServerGrantUrl:         "http://mockauth",
			AuthServerIntrospectionUrl: "http://mockauth",
			StreamSecret:               streamSecretBase64,
			AdminKey:                   streamSecret.Result,                     // re-using for local cluster
			WebhookUrl:                 "https://end3sf6r22xva.x.pipedream.net", // using request bin for now
			Hostname:                   "peer.fynbos.test",
			PublicHost:                 "http://peer", // using peer as coredns will find peer
		})
		if err != nil {
			return err
		}

		if err := mockstaticresponseservers.DeployMockGnapServer(ctx, "mockauth"); err != nil {
			return err
		}

		return nil
	})
}
