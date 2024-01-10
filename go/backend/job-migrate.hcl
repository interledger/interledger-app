job "backend-migrate" {
  datacenters = ["dc1"]
  type = "batch"

  group "backend-migrate" {
    count = 1


    network {
      mode = "bridge"
    }

    service {
      name = "backend-migrate"
      tags = [
        "traefik.enable=false",
      ]

      connect {
        sidecar_service {
          proxy {
            upstreams {
              destination_name = "postgres"
              local_bind_port = 5432
            }
            upstreams {
              destination_name = "temporal"
              local_bind_port = 7233
            }
          }
          # Super NB: Don't let traefik route to the connect proxy ingress
          # for this service (it doesn't speak mTLS, so isn't mesh friendly)
          # This works fine for now, as we run an LB in the local network
          # but wouldn't work if we eg ran a load balancer for a service
          # outside
          tags = [
            "traefik.enable=false"
          ]
        }
      }
    }

    task "backend-migrate" {
      driver = "docker"
      config {
        image   = "backend:local"
        command = "migrate"
      }

      env {
        FYNBOS_ENV            = "local"
        ENV_FILE              = ""
        TWILIO_ACCOUNT_TOKEN  = "test"
        ZENDESK_TOKEN         = "test"
        PUSHER_ADDR           = "https://91988d6075551d29760a:6b0d52daa5caf08e4b81@api-eu.pusher.com/apps/1538039"
        AUTHORISATION_PORT    = 8082
        OPEN_PAYMENTS_PORT    = 8081
        TWITTER_CLIENT_SECRET = "hKpTMGuoluBTVippudUJniNgw_yelIUHSvnuqiheyqsGd7MZ6Y"
        TWITTER_REDIRECT_URL  = "https://fynbos.test/connect/twitter"
        TWITTER_CLIENT_ID     = "Q2pzaXpDN0VPN29MU09VQ0tNYmo6MTpjaQ"
        DB_URL_WITH_CERTS           = "postgres://postgres:password@localhost:5432/backend?sslmode=disable"
        DB_URL                      = "postgres://postgres:password@localhost:5432/backend?sslmode=disable"
        PACIOLI_DB_URL_WITH_CERTS   = "postgres://postgres:password@localhost:5432/pacioli?sslmode=disable"
        PACIOLI_DB_URL              = "postgres://postgres:password@localhost:5432/pacioli?sslmode=disable"
        KRATOS_URL                  = "http://kratos-public.kratos"
        KRATOS_ADMIN_URL            = "http://kratos-admin.kratos"
        LOG_LEVEL                   = "info"
        USD_LEDGER_ID               = "1"
        NOOP_EQUITY_ACCOUNT_ID      = "43d4b2bd-e29b-4a63-9aa8-7990776c714e"
        PACIOLI_URL                 = "pacioli.pacioli:443"
        GOOGLE_OUATH2_CLIENT_ID     = "google_oauth"
        RAFIKI_GRAPHQL_URL          = "http://rafiki.rafiki/graphql"
        TEMPORAL_URL                = "localhost:7233"
        TWILIO_ACCOUNT_SID          = "SK021f793191208ba69c3bea87dd426085"
        TWILIO_SERVICE_SID          = "VA8af4e130da63b9fac4c042acbc33a267"
        ZENDESK_USER                = "matt@fynbos.dev"
        OTEL_EXPORTER_OTLP_ENDPOINT = "grpc://api.honeycomb.io:443"
        OTEL_EXPORTER_OTLP_HEADERS  = "x-honeycomb-team=7Qskhns7Dc7wgazrDe6yZD"
        OTEL_SERVICE_NAME           = "backend"
      }
    }
  }
}