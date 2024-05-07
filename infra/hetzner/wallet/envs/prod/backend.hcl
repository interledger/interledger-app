job "backend" {
  datacenters = ["dc1"]
  type = "service"
  namesapce = "prod"

  
  group "backend" {
    count = 1

    network {
      mode = "bridge"
      port "grpc" {
        to = 8443
      }

      port "http" {
        to = 8080
      }

      port "admin" {
        to = 8448
      }
    }

    service {
      name = "backend-grpc"
      port = 8443
      connect {
        sidecar_service {}
      }
    }

    service {
      name = "backend-admin"
      port = 8448
      connect {
        sidecar_service {}
      }
    }

    service {
      name = "backend-http"
      port = 8080
      connect {
        sidecar_service {}
      }
      tags = [
        "traefik.enable=true",
        "traefik.http.routers.backend.rule=Host(`fynbos.app`) && PathPrefix(`/webhooks`)",
        "traefik.http.routers.backend.rule=Host(`fynbos.me`)",
      ]
    }

    service {
      connect {
        sidecar_service {
          proxy {
            upstreams {
              destination_name = "postgres"
              local_bind_port  = 5432
            }
            upstreams {
              destination_name = "kratos"
              local_bind_port = 4433
            }
            upstreams {
              destination_name = "kratos-admin"
              local_bind_port = 4434
            }
            upstreams {
              destination_name = "temporal"
              local_bind_port = 7233
            }
            upstreams {
              destination_name = "rafiki-auth-admin"
              local_bind_port = 3003
            }
            upstreams {
              destination_name = "rafiki-backend-admin"
              local_bind_port = 3001
            }
          }
        }
      }
    }

    task "backend-migrations" {
      driver = "docker"
      kill_timeout = "120s"

      config {
        image = "localhost:5002/backend"
        args= ["migrate"]
      }

      lifecycle {
        hook = "prestart"
        sidecar = false
      }

      template {
        data = <<EOH
          DB_URL_WITH_CERTS="postgres://{{with secret "database-prod/creds/backend"}}{{.Data.data.username}}:{{.Data.password}}{{end}}@localhost:5432/backend?sslmode=disable"
          PACIOLI_DB_URL_WITH_CERTS="postgres://{{with secret "database-prod/creds/pacioli"}}{{.Data.data.username}}:{{.Data.password}}{{end}}@localhost:5432/pacioli?sslmode=disable"
          KRATOS_URL="http://localhost:4433"
          LOG_LEVEL="info"
        EOH

        destination = "secrets/file.env"
        env         = true
      }
    }

    task "backend" {
      driver = "docker"

      config {
        image = "localhost:5002/backend"
        args= ["start"]
      }

      template {
        data = <<EOH
          FYNBOS_ENV=prod
          {{with secret "database-prod/creds/backend"}}
          DB_URL_WITH_CERTS="postgres://{{.Data.data.username}}:{{.Data.password}}@localhost:5432/backend?sslmode=disable"
          DB_URL="postgres://{{.Data.data.username}}:{{.Data.password}}@localhost:5432/backend?sslmode=disable"
          {{end}}

          {{with secret "database-prod/creds/pacioli"}}
          PACIOLI_DB_URL_WITH_CERTS="postgres://{{.Data.data.username}}:{{.Data.password}}@localhost:5432/pacioli?sslmode=disable"
          PACIOLI_DB_URL="postgres://{{.Data.data.username}}:{{.Data.password}}@localhost:5432/pacioli?sslmode=disable"
          {{end}}
          KRATOS_URL="http://localhost:4433"
          KRATOS_ADMIN_URL="http://localhost:4434"
          LOG_LEVEL="info"
          USD_LEDGER_ID="1"
          NOOP_EQUITY_ACCOUNT_ID="43d4b2bd-e29b-4a63-9aa8-7990776c714e"
          GOOGLE_OUATH2_CLIENT_ID="google_oauth"
          RAFIKI_GRAPHQL_URL="http://rafiki.rafiki/graphql"
          TEMPORAL_URL="localhost:7233"
          ENV_FILE=
          TWILIO_ACCOUNT_SID: "SKafd3d83b760b275b052cb4d2cad07749"
          TWILIO_SERVICE_SID: "VAfed340e6a933e63f95f3ab6058d7805b"
          ZENDESK_USER="matt@fynbos.dev"
          ZENDESK_TOKEN=test
          OTEL_EXPORTER_OTLP_ENDPOINT="grpc://api.honeycomb.io:443"
          OTEL_EXPORTER_OTLP_HEADERS: "x-honeycomb-team=oTtj00yo3Le8WofuInVwFB"
          OTEL_SERVICE_NAME="backend"
          ADMIN_POLICY_AUD: "8fba89bfeac98b04e09af97166118ae2339e0f9de743b2c980b2cb6a0c0e3878"
          ADMIN_TEAM_DOMAIN: "https://fynbos.cloudflareaccess.com"
          AUTHORISATION_PORT: "8082"
          OPEN_PAYMENTS_PORT: "8081"
          VAULT_ADDR: "https://vault1.fynbos.cloud:8200"
          VAULT_AUTH_PATH: "k8s-prod-use2"
          VAULT_TRANSIT_ENGINE_PATH: "transit/k8s-prod-use2/backend/backend"
          TWITTER_REDIRECT_URL: "https://fynbos.app/connect/twitter"
          TWITTER_CLIENT_ID: "Q2pzaXpDN0VPN29MU09VQ0tNYmo6MTpjaQ"
          DISCORD_REDIRECT_URL: "https://fynbos.app/connect/discord"
          ASTRA_CODE_EXCHANGE_REDIRECT: "https://enxt6s49y9jsd.x.pipedream.net/"
          RAFIKI_BACKEND_GRAPHQL_URL=http://127.0.0.1:3001/graphql
          RAFIKI_AUTH_GRAPHQL_URL=http://127.0.0.1:3003/graphql

          {{with secret "kv/data/prod/backend/config"}}
          TWILIO_ACCOUNT_TOKEN={{.Data.data.TWILIO_ACCOUNT_TOKEN}}
          ZENDESK_TOKEN={{.Data.data.ZENDESK_TOKEN}}
          SENDGRID_API_KEY={{.Data.data.SENDGRID_API_KEY}}
          SMARTY_AUTH_ID={{.Data.data.SMARTY_AUTH_ID}}
          SMARTY_AUTH_TOKEN={{.Data.data.SMARTY_AUTH_TOKEN}}
          PUSHER_ADDR={{.Data.data.PUSHER_ADDR}}
          SEGMENT_KEY={{.Data.data.SEGMENT_KEY}}
          PERSONA_TOKEN={{.Data.data.PERSONA_TOKEN}}
          PERSONA_WEBHOOK_TOKEN={{.Data.data.PERSONA_WEBHOOK_TOKEN}}
          SLACK_TOKEN={{.Data.data.SLACK_TOKEN}}
          TWITTER_CLIENT_SECRET={{.Data.data.TWITTER_CLIENT_SECRET}}
          CDN_KEY={{.Data.data.CDN_KEY}}
          BASISTHEORY_API_KEY={{.Data.data.BASISTHEORY_API_KEY}}
          TWITTER_BEARER_TOKEN={{.Data.data.TWITTER_BEARER_TOKEN}}
          SENTRY_DSN={{.Data.data.SENTRY_DSN}}
          DISCORD_CLIENT_SECRET={{.Data.data.DISCORD_CLIENT_SECRET}}
          DISCORD_CLIENT_ID={{.Data.data.DISCORD_CLIENT_ID}}
          SLACK_CLIENT_ID={{.Data.data.SLACK_CLIENT_ID}}
          SLACK_CLIENT_SECRET={{.Data.data.SLACK_CLIENT_SECRET}}
          SLACK_SIGNING_SECRET={{.Data.data.SLACK_SIGNING_SECRET}}
          SENTRY_DSN=.{{.Data.data.SENTRY_DSN}}
          XAGO_API_SECRET={{.Data.data.XAGO_API_SECRET}}
          XAGO_API_PUBLIC_KEY={{.Data.data.XAGO_API_PUBLIC_KEY}}
          ASTRA_WEBHOOK_BEARER_TOKEN={{.Data.data.ASTRA_WEBHOOK_BEARER_TOKEN}}
          ASTRA_CLIENT_ID={{.Data.data.ASTRA_CLIENT_ID}}
          ASTRA_CLIENT_SECRET={{.Data.data.ASTRA_CLIENT_SECRET}}
          ASTRA_ACCOUNT_NUMBER={{.Data.data.ASTRA_ACCOUNT_NUMBER}}
          ASTRA_ROUTING_NUMBER={{.Data.data.ASTRA_ROUTING_NUMBER}}
          PTI_JWK={{.Data.data.PTI_JWK}}
          PTI_CLIENT_ID={{.Data.data.PTI_CLIENT_ID}}
          GATEHUB_APP_ID={{.Data.data.GATEHUB_APP_ID}}
          GATEHUB_SECRET={{.Data.data.GATEHUB_SECRET}}
          GATEHUB_WEBHOOK_SECRET={{.Data.data.GATEHUB_WEBHOOK_SECRET}}
          {{end}}
        EOH

        destination = "secrets/file.env"
        env         = true
      }
    }

  }
}
