job "backend" {
  datacenters = ["dc1"]
  type = "service"

  
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
          DB_URL_WITH_CERTS="postgres://postgres:password@localhost:5432/backend?sslmode=disable"
          PACIOLI_DB_URL_WITH_CERTS="postgres://postgres:password@localhost:5432/pacioli?sslmode=disable"
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
          FYNBOS_ENV=local
          DB_URL_WITH_CERTS="postgres://postgres:password@localhost:5432/backend?sslmode=disable"
          DB_URL="postgres://postgres:password@localhost:5432/backend?sslmode=disable"
          PACIOLI_DB_URL_WITH_CERTS="postgres://postgres:password@localhost:5432/pacioli?sslmode=disable"
          PACIOLI_DB_URL="postgres://postgres:password@localhost:5432/pacioli?sslmode=disable"
          KRATOS_URL="http://localhost:4433"
          KRATOS_ADMIN_URL="http://localhost:4434"
          LOG_LEVEL="info"
          USD_LEDGER_ID="1"
          NOOP_EQUITY_ACCOUNT_ID="43d4b2bd-e29b-4a63-9aa8-7990776c714e"
          GOOGLE_OUATH2_CLIENT_ID="google_oauth"
          RAFIKI_GRAPHQL_URL="http://rafiki.rafiki/graphql"
          TEMPORAL_URL="localhost:7233"
          ENV_FILE=
          TWILIO_ACCOUNT_SID="SK021f793191208ba69c3bea87dd426085"
          TWILIO_SERVICE_SID="VA8af4e130da63b9fac4c042acbc33a267"
          TWILIO_ACCOUNT_TOKEN=test
          ZENDESK_USER="matt@fynbos.dev"
          ZENDESK_TOKEN=test
          OTEL_EXPORTER_OTLP_ENDPOINT="grpc://api.honeycomb.io:443"
          OTEL_EXPORTER_OTLP_HEADERS="x-honeycomb-team=7Qskhns7Dc7wgazrDe6yZD"
          OTEL_SERVICE_NAME="backend"
        EOH

        destination = "secrets/file.env"
        env         = true
      }
    }

  }
}
