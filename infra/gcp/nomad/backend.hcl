variable "image_hash" {
  type = string
}

job "backend" {
  datacenters = ["dc1"]
  type = "service"
  namespace = "dev"

  
  group "backend" {
    count = 1

    update {
      max_parallel     = 1
      canary           = 1
      auto_revert      = true
      auto_promote     = true
      health_check     = "checks"
    }

    network {
      mode = "bridge"
      port "http" {
        to = 8080
      }
    }

    service {
      name = "dev-backend-grpc"
      port = 8443
      connect {
        sidecar_service {}
      }
    }

    service {
      name = "dev-backend-admin"
      port = 8448
      connect {
        sidecar_service {}
      }
    }

    service {
      name = "dev-backend-http"
      port = 8080

      connect {
        sidecar_service {}
      }
    }

    service {
      name = "dev-backend-openpayments"
      port = "http"
      
      tags = [
        "traefik.enable=true",
        "traefik.http.routers.dev-backend-openpayments.rule=Host(`eu1.fynbos.me`)"
      ]
    }

    service {
      name = "dev-backend"
      port = "http"

      tags = [
        "traefik.enable=true",
        "traefik.http.routers.dev-backend.rule=Host(`eu1.fynbos.dev`) && PathPrefix(`/webhooks`)"
      ]

      check {    
        type     = "http"
        port     = "http"
        path     = "/healthz"
        interval = "5s"
        timeout  = "2s"
      }

      connect {
        sidecar_service {
          tags = [
            "treafik.enable=false"
          ]

          proxy {
            upstreams {
              destination_name = "dev-postgres"
              local_bind_port  = 5432
            }
            upstreams {
              destination_name = "dev-kratos"
              local_bind_port = 4433
            }
            upstreams {
              destination_name = "dev-kratos-admin"
              local_bind_port = 4434
            }
            upstreams {
              destination_name = "dev-temporal-frontend"
              local_bind_port = 7233
            }
            upstreams {
              destination_name = "dev-rafiki-auth-admin"
              local_bind_port = 3003
            }
            upstreams {
              destination_name = "dev-rafiki-backend-admin"
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
        image = format("registry.gitlab.com/fynbos/fynbos/backend:%s", var.image_hash)
        args= ["migrate"]
      }

      lifecycle {
        hook = "prestart"
        sidecar = false
      }

      vault {}

      template {
        data = <<EOH
          DB_URL_WITH_CERTS="postgres://{{with secret "database-dev/static-creds/backend"}}{{.Data.username}}:{{.Data.password}}{{end}}@localhost:5432/backend?sslmode=disable"
          PACIOLI_DB_URL_WITH_CERTS="postgres://{{with secret "database-dev/static-creds/pacioli"}}{{.Data.username}}:{{.Data.password}}{{end}}@localhost:5432/pacioli?sslmode=disable"
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
        image = format("registry.gitlab.com/fynbos/fynbos/backend:%s", var.image_hash)
        args= ["start"]
      }

      vault {}

      template {
        data = <<EOH
          FYNBOS_ENV=dev
          {{with secret "database-dev/static-creds/backend"}}
          DB_URL_WITH_CERTS="postgres://{{.Data.username}}:{{.Data.password}}@localhost:5432/backend?sslmode=disable"
          DB_URL="postgres://{{.Data.username}}:{{.Data.password}}@localhost:5432/backend?sslmode=disable"
          {{end}}

          {{with secret "database-dev/static-creds/pacioli"}}
          PACIOLI_DB_URL_WITH_CERTS="postgres://{{.Data.username}}:{{.Data.password}}@localhost:5432/pacioli?sslmode=disable"
          PACIOLI_DB_URL="postgres://{{.Data.username}}:{{.Data.password}}@localhost:5432/pacioli?sslmode=disable"
          {{end}}
          KRATOS_URL="http://localhost:4433"
          KRATOS_ADMIN_URL="http://localhost:4434"
          LOG_LEVEL="info"
          USD_LEDGER_ID="1"
          NOOP_EQUITY_ACCOUNT_ID="43d4b2bd-e29b-4a63-9aa8-7990776c714e"
          GOOGLE_OUATH2_CLIENT_ID="google_oauth"
          TEMPORAL_URL="localhost:7233"
          ENV_FILE=
          TWILIO_ACCOUNT_SID="SKafd3d83b760b275b052cb4d2cad07749"
          TWILIO_SERVICE_SID="VAfed340e6a933e63f95f3ab6058d7805b"
          ZENDESK_USER="matt@fynbos.dev"
          ZENDESK_TOKEN=test
          OTEL_EXPORTER_OTLP_ENDPOINT="grpc://api.honeycomb.io:443"
          OTEL_EXPORTER_OTLP_HEADERS="x-honeycomb-team=7Qskhns7Dc7wgazrDe6yZD"
          OTEL_SERVICE_NAME="backend"
          ADMIN_POLICY_AUD="8fba89bfeac98b04e09af97166118ae2339e0f9de743b2c980b2cb6a0c0e3878"
          ADMIN_TEAM_DOMAIN="https://fynbos.cloudflareaccess.com"
          AUTHORISATION_PORT="8082"
          OPEN_PAYMENTS_PORT="8081"
          VAULT_ADDR="http://172.26.64.1:8200"
          VAULT_TRANSIT_ENGINE_PATH="transit/dev/backend"
          TWITTER_REDIRECT_URL="https://fynbos.app/connect/twitter"
          TWITTER_CLIENT_ID="Q2pzaXpDN0VPN29MU09VQ0tNYmo6MTpjaQ"
          DISCORD_REDIRECT_URL="https://fynbos.app/connect/discord"
          ASTRA_CODE_EXCHANGE_REDIRECT="https://enxt6s49y9jsd.x.pipedream.net/"
          RAFIKI_USD_ASSET="80d80585-5341-413a-acaf-169779b4642c"
          FYNBOS_ADDRESS_STATE="CA"
          FYNBOS_ADDRESS_CITY="San Francisco"
          FYNBOS_ADDRESS_ZIPCODE="94103"
          FYNBOS_ADDRESS_COUNTRY="840"
          FYNBOS_ADDRESS_COUNTY="038"
          FYNBOS_ADDRESS_LINE1="785 Market Street"
          RAFIKI_BACKEND_GRAPHQL_URL=http://127.0.0.1:3001/graphql
          RAFIKI_AUTH_GRAPHQL_URL=http://127.0.0.1:3003/graphql
          PTI_PUBLIC_KEY_JWK='{"e":"AQAB","kid":"84d1a616-db99-47dd-b690-70742b87aa4e","kty":"RSA","n":"1EWed921xDloYd1SbAkUsHbPqeKtGIOAmZhT4aYrT3F1-rugltXSSugO7SlLuUlH1k1a2tJyK3TtCbuSHE11Y8SGQeRcUFvLISz6XqBgAA59NOj_6gIMuE1L4MM5nEfbFHxq6YqDMDO5hZ5_P-2BMS-dtxx9VoPgUJC_OuglDvckpG3xxqITIbhfGsDe2_KolEHRkW8ozGNfR2kAKpuJFTqJTUg7jFDmNq_zgUA6_13d9F_tWN6uyW5X0gN9UjpEE86uFh49ScjwVthcRamiqk-PmrGmgoFrrplszf6bD7QScytecrag92Ls1kmX00BHnV0MBsh-Bk8vE8u88MNT0dfj5OLMIhqS014najHSMnXdTZg2iav9-r-iAQ5vHnfMGupR3jBttGlFWp72cKZUqA_Tu0m4efxiTzWJk10rDOdLOBsQBkDmMC72X-FmfxYXscr4vrDRvbGAfEFoceanYg8fQbr0A4mQtLZlAAsmLBBuLHvw3MPpW2eD1H8wicm2i0J8_EHdhwp3u-83rtkVhowmjE9hav9lpjiFmSNGFYjR8ddxmRNM-vDc4o165frPnGpAhqTnnDelZcNT_CFkgwQ4H38J4ciuB50pM8meoCbapKunK9DqH0gs2iwpMUcq3B0pWJ27nvtnvujiVXTWCP9A8Dk4J-LQUU-TquQy8k0"}'

          {{with secret "kv/data/dev/backend/config"}}
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
          DISCORD_CLIENT_SECRET={{.Data.data.DISCORD_CLIENT_SECRET}}
          DISCORD_CLIENT_ID={{.Data.data.DISCORD_CLIENT_ID}}
          SLACK_CLIENT_ID={{.Data.data.SLACK_CLIENT_ID}}
          SLACK_CLIENT_SECRET={{.Data.data.SLACK_CLIENT_SECRET}}
          SLACK_SIGNING_SECRET={{.Data.data.SLACK_SIGNING_SECRET}}
          XAGO_API_SECRET={{.Data.data.XAGO_API_SECRET}}
          XAGO_API_PUBLIC_KEY={{.Data.data.XAGO_API_PUBLIC_KEY}}
          ASTRA_WEBHOOK_BEARER_TOKEN={{.Data.data.ASTRA_WEBHOOK_BEARER_TOKEN}}
          ASTRA_CLIENT_ID={{.Data.data.ASTRA_CLIENT_ID}}
          ASTRA_CLIENT_SECRET={{.Data.data.ASTRA_CLIENT_SECRET}}
          PTI_JWK='{{.Data.data.PTI_JWK}}'
          PTI_CLIENT_ID={{.Data.data.PTI_CLIENT_ID}}
          GATEHUB_APP_ID={{.Data.data.GATEHUB_APP_ID}}
          GATEHUB_SECRET={{.Data.data.GATEHUB_SECRET}}
          GATEHUB_WEBHOOK_SECRET={{.Data.data.GATEHUB_WEBHOOK_SECRET}}
          CHIMONEY_WEBHOOK_SECRET={{.Data.data.CHIMONEY_WEBHOOK_SECRET}}
          CHIMONEY_TOKEN={{.Data.data.CHIMONEY_TOKEN}}
          {{end}}
        EOH

        destination = "secrets/file.env"
        env         = true
      }
    }

    task "worker" {
      driver = "docker"

      config {
        image = format("registry.gitlab.com/fynbos/fynbos/backend:%s", var.image_hash)
        args= ["worker"]
      }

      vault {}

      template {
        data = <<EOH
          {{with secret "database-dev/static-creds/backend"}}
          DB_URL_WITH_CERTS="postgres://{{.Data.username}}:{{.Data.password}}@localhost:5432/backend?sslmode=disable"
          DB_URL="postgres://{{.Data.username}}:{{.Data.password}}@localhost:5432/backend?sslmode=disable"
          {{end}}

          {{with secret "database-dev/static-creds/pacioli"}}
          PACIOLI_DB_URL_WITH_CERTS="postgres://{{.Data.username}}:{{.Data.password}}@localhost:5432/pacioli?sslmode=disable"
          PACIOLI_DB_URL="postgres://{{.Data.username}}:{{.Data.password}}@localhost:5432/pacioli?sslmode=disable"
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
          TWILIO_ACCOUNT_SID="SKafd3d83b760b275b052cb4d2cad07749"
          TWILIO_SERVICE_SID="VAfed340e6a933e63f95f3ab6058d7805b"
          ZENDESK_USER="matt@fynbos.dev"
          ZENDESK_TOKEN=test
          OTEL_EXPORTER_OTLP_ENDPOINT="grpc://api.honeycomb.io:443"
          OTEL_EXPORTER_OTLP_HEADERS="x-honeycomb-team=oTtj00yo3Le8WofuInVwFB"
          OTEL_SERVICE_NAME="backend"
          ADMIN_POLICY_AUD="8fba89bfeac98b04e09af97166118ae2339e0f9de743b2c980b2cb6a0c0e3878"
          ADMIN_TEAM_DOMAIN="https://fynbos.cloudflareaccess.com"
          AUTHORISATION_PORT="8082"
          OPEN_PAYMENTS_PORT="8081"
          VAULT_ADDR="http://172.26.64.1:8200"
          VAULT_TRANSIT_ENGINE_PATH="transit/dev/backend"
          TWITTER_REDIRECT_URL="https://fynbos.app/connect/twitter"
          TWITTER_CLIENT_ID="Q2pzaXpDN0VPN29MU09VQ0tNYmo6MTpjaQ"
          DISCORD_REDIRECT_URL="https://fynbos.app/connect/discord"
          ASTRA_CODE_EXCHANGE_REDIRECT="https://enxt6s49y9jsd.x.pipedream.net/"
          RAFIKI_USD_ASSET="80d80585-5341-413a-acaf-169779b4642c"
          FYNBOS_ADDRESS_STATE="CA"
          FYNBOS_ADDRESS_CITY="San Francisco"
          FYNBOS_ADDRESS_ZIPCODE="94103"
          FYNBOS_ADDRESS_COUNTRY="840"
          FYNBOS_ADDRESS_COUNTY="038"
          FYNBOS_ADDRESS_LINE1="785 Market Street"
          RAFIKI_BACKEND_GRAPHQL_URL=http://127.0.0.1:3001/graphql
          RAFIKI_AUTH_GRAPHQL_URL=http://127.0.0.1:3003/graphql

          {{with secret "kv/data/dev/backend/config"}}
          TWILIO_ACCOUNT_TOKEN={{.Data.data.TWILIO_ACCOUNT_TOKEN}}
          ZENDESK_TOKEN={{.Data.data.ZENDESK_TOKEN}}
          SENDGRID_API_KEY={{.Data.data.SENDGRID_API_KEY}}
          SMARTY_AUTH_ID={{.Data.data.SMARTY_AUTH_ID}}
          SMARTY_AUTH_TOKEN={{.Data.data.SMARTY_AUTH_TOKEN}}
          PUSHER_ADDR={{.Data.data.PUSHER_ADDR}}
          SEGMENT_KEY={{.Data.data.SEGMENT_KEY}}
          SLACK_TOKEN={{.Data.data.SLACK_TOKEN}}
          BASISTHEORY_API_KEY={{.Data.data.BASISTHEORY_API_KEY}}
          PERSONA_TOKEN={{.Data.data.PERSONA_TOKEN}}
          PERSONA_WEBHOOK_TOKEN={{.Data.data.PERSONA_WEBHOOK_TOKEN}}
          TWITTER_CLIENT_SECRET={{.Data.data.TWITTER_CLIENT_SECRET}}
          TWITTER_BEARER_TOKEN={{.Data.data.TWITTER_BEARER_TOKEN}}
          CDN_KEY={{.Data.data.CDN_KEY}}
          XAGO_API_SECRET={{.Data.data.XAGO_API_SECRET}}
          XAGO_API_PUBLIC_KEY={{.Data.data.XAGO_API_PUBLIC_KEY}}
          PTI_JWK='{{.Data.data.PTI_JWK}}'
          PTI_CLIENT_ID={{.Data.data.PTI_CLIENT_ID}}
          ASTRA_CLIENT_ID={{.Data.data.ASTRA_CLIENT_ID}}
          ASTRA_CLIENT_SECRET={{.Data.data.ASTRA_CLIENT_SECRET}}
          GATEHUB_APP_ID={{.Data.data.GATEHUB_APP_ID}}
          GATEHUB_SECRET={{.Data.data.GATEHUB_SECRET}}
          CHIMONEY_WEBHOOK_SECRET={{.Data.data.CHIMONEY_WEBHOOK_SECRET}}
          CHIMONEY_TOKEN={{.Data.data.CHIMONEY_TOKEN}}
          {{end}}

          FYNBOS_ENV=dev
          ENV_FILE=
          TWILIO_ACCOUNT_SID="SKafd3d83b760b275b052cb4d2cad07749"
          TWILIO_SERVICE_SID="VAfed340e6a933e63f95f3ab6058d7805b"
          ZENDESK_USER="matt@fynbos.dev"
        EOH

        destination = "secrets/file.env"
        env         = true
      }
    }

  }
}
