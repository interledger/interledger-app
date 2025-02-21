job "backend" {
  datacenters = ["dc1"]
  type = "service"
  
  group "backend" {
    count = 1

    network {
      mode = "bridge"
      port "http" {
        to = 8080
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
    }

    volume "go" {
      type = "host"
      read_only = false
      source = "go"
    }

    service {
      name = "backend-openpayments"
      port = "http"
      
      tags = [
        "traefik.enable=true",
        "traefik.http.routers.backend-openpayments.rule=Host(`local.fynbos.me`)"
      ]
    }

    service {
      name = "backend"
      port = "http"

      tags = [
        "traefik.enable=true",
        "traefik.http.routers.backend.rule=Host(`wallet.fynbos.test`) && PathPrefix(`/webhooks`)"
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
            "traefik.enable=false"
          ]

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
              destination_name = "mockbos"
              local_bind_port = 9080
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

    task "mod-download" {
      driver = "docker"

      config {
        image = "localhost:5002/backend"
        entrypoint = ["/bin/sh"]
        args =  ["-c", "go mod download && go install github.com/air-verse/air@latest"]
        volumes = [
          "/home/vagrant/fynbos/go:/build",
          "${NOMAD_ALLOC_DIR}/go:/go"
        ]
      }

      lifecycle {
        hook = "prestart"
        sidecar = false
      }
    }

    task "backend-migrations" {
      driver = "docker"

      config {
        image = "localhost:5002/backend"
        entrypoint = ["/go/bin/air"]
        args =  ["--build.poll", "true", "--build.include_ext", "hcl", "--build.cmd", "go build -o /local/main /build/backend/main.go", "--build.bin", "/local/main migrate"]
        volumes = [
          "/home/vagrant/fynbos/go:/build",
          "${NOMAD_ALLOC_DIR}/go:/go",
          "${NOMAD_ALLOC_DIR}/go/cache:/root/.cache/go-build"
        ]
      }

      env {
          FYNBOS_ENV = "local"
          DB_URL_WITH_CERTS = "postgres://postgres:password@127.0.0.1:5432/backend?sslmode=disable"
          PACIOLI_DB_URL_WITH_CERTS = "postgres://postgres:password@127.0.0.1:5432/pacioli?sslmode=disable"
          KRATOS_URL = "http://127.0.0.1:4433"
          LOG_LEVEL = "info"
      }

      resources {
        memory_max = 750
      }
    }

    task "backend" {
      driver = "docker"

      config {
        image = "localhost:5002/backend"
        entrypoint = ["/go/bin/air"]
        args =  ["--build.poll", "true", "--build.cmd", "go build -o /local/main /build/backend/main.go", "--build.bin", "/local/main dev"]
        volumes = [
          "/home/vagrant/fynbos/go:/build",
          "${NOMAD_ALLOC_DIR}/go:/go",
          "${NOMAD_ALLOC_DIR}/go/cache:/root/.cache/go-build"
        ]
      }

      env {
        FYNBOS_ENV = "local"
        DB_URL_WITH_CERTS = "postgres://postgres:password@localhost:5432/backend?sslmode=disable"
        DB_URL = "postgres://postgres:password@localhost:5432/backend?sslmode=disable"
        PACIOLI_DB_URL_WITH_CERTS = "postgres://postgres:password@localhost:5432/pacioli?sslmode=disable"
        PACIOLI_DB_URL = "postgres://postgres:password@localhost:5432/pacioli?sslmode=disable"
        KRATOS_URL = "http://localhost:4433"
        KRATOS_ADMIN_URL = "http://localhost:4434"
        LOG_LEVEL = "info"
        USD_LEDGER_ID = "1"
        NOOP_EQUITY_ACCOUNT_ID = "43d4b2bd-e29b-4a63-9aa8-7990776c714e"
        GOOGLE_OUATH2_CLIENT_ID = "google_oauth"
        RAFIKI_GRAPHQL_URL = "http://rafiki.rafiki/graphql"
        TEMPORAL_URL = "localhost:7233"
        ENV_FILE = ""
        TWILIO_ACCOUNT_SID = "SK021f793191208ba69c3bea87dd426085"
        TWILIO_SERVICE_SID = "VA8af4e130da63b9fac4c042acbc33a267"
        TWILIO_ACCOUNT_TOKEN = "test"
        ZENDESK_USER = "matt@fynbos.dev"
        ZENDESK_TOKEN = "test"
        OTEL_EXPORTER_OTLP_ENDPOINT = "grpc://api.honeycomb.io:443"
        OTEL_EXPORTER_OTLP_HEADERS = "x-honeycomb-team=7Qskhns7Dc7wgazrDe6yZD"
        OTEL_SERVICE_NAME = "backend"
      }

      resources {
        memory_max = 750
      }
    }

  }
}
