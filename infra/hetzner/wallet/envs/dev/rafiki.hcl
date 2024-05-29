job "rafiki" {
  datacenters = ["dc1"]
  type = "service"
  namesapce = "dev"

  group "backend" {
    count = 1

    network {
      mode = "bridge"
      port "open-payments" {
        to = 80
      }

      port "auth" {
        to = 3006
      }

      port "introspection" {
        to = 3007
      }

      port "frontend" {
        to = 3010
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
              destination_name = "redis"
              local_bind_port = 6379
            }
            upstreams {
              destination_name = "backend-http"
              local_bind_port = 8080
            }
          }
        }
      }
    }

    service {
      name = "rafiki-backend-admin"
      port = 3001
      connect {
        sidecar_service {}
      }
    }

    service {
      name = "rafiki-connector"
      port = 3002
      tags = [
        "traefik.enable=true",
        "traefikl.http.middlewares.rafiki-connector-stripprefix.stripprefix=/ilp",
        "traefik.http.routers.rafiki-connector.rule=Host(`eu1.fynbos.me`) && Path(`/ilp`)",
        "traefikl.http.routers.rafiki-connector.middlewares=rafiki-connector-stripprefix@consulcatalog"
      ]
    }

    service {
      name = "rafiki-openpayments"
      port = 80
      tags = [
        "traefik.enable=true",
        "traefik.http.routers.rafiki-openpayments.rule=Host(`eu1.fynbos.me`) && PathPrefix(`/incoming-payments`)",
        "traefik.http.routers.rafiki-openpayments.rule=Host(`eu1.fynbos.me`) && PathPrefix(`/outgoing-payments`)",
        "traefik.http.routers.rafiki-openpayments.rule=Host(`eu1.fynbos.me`) && PathPrefix(`/quotes`)",
        "traefik.http.routers.rafiki-openpayments.rule=Host(`eu1.fynbos.me`) && PathPrefix(`/.well-known`)"
      ]
    }

    service {
      name = "rafiki-auth"
      port = 3006
      tags = [
        "traefik.enable=true",
        "traefikl.http.middlewares.rafiki-auth-stripprefix.stripprefix=/gnap",
        "traefik.http.routers.rafiki-auth.rule=Host(`eu1.fynbos.me`) && Path(`/gnap`)",
        "traefik.http.routers.rafiki-auth.rule=Host(`eu1.fynbos.me`) && PathPrefix(`/token`)",
        "traefik.http.routers.rafiki-auth.rule=Host(`eu1.fynbos.me`) && PathPrefix(`/interact`)",
        "traefik.http.routers.rafiki-auth.rule=Host(`eu1.fynbos.me`) && PathPrefix(`/continue`)",
        "traefik.http.routers.rafiki-auth.middlewars=rafiki-auth-stripprefix@consulcatalog"
      ]
    }

    service {
      name = "rafiki-auth-admin"
      port = 3003
      connect {
        sidecar_service {}
      }
    }

    service {
      name = "rafiki-frontend"
      port = "frontend"
      tags = [
        "traefik.enable=true",
        "traefik.http.routers.rafiki-frontend.rule=Host(`rafiki-dev.mgnt.fynbos.dev`)"
      ]
    }

    task "auth" {
      driver = "docker"

      config {
        image = "localhost:5002/rafiki-auth"
      }

      template {
        data = <<EOF
        AUTH_DATABASE_URL="postgres://{{with secret "database-dev/creds/rafiki_auth"}}{{.Data.data.username}}:{{.Data.password}}{{end}}@localhost:5432/rafiki_auth?sslmode=disable"
        {{with secret "kv/data/dev/rafiki_auth/config"}}
        IDENTITY_SERVER_SECRET={{.Data.data.IDENTITY_SERVER_SECRET}}
        COOKIE_KEY: ""
        {{end}}
        EOF
        destination - "secrets/file.env"
        env = true
      }

      template {
        data = <<EOF
        ACCESS_TOKEN_DELETION_DAYS: "30"
        ACCESS_TOKEN_EXPIRY_SECONDS: "600"
        ADMIN_PORT: "3003"
        AUTH_PORT: "3006"
        AUTH_SERVER_DOMAIN: "http://rafiki-rafiki-auth:3006"
        DATABASE_CLEANUP_WORKERS: "1"
        IDENTITY_SERVER_DOMAIN: "http://cloud-nine-wallet/idp"
        INCOMING_PAYMENT_INTERACTION: "false"
        QUOTE_INTERACTION: "false"
        INTROSPECTION_PORT: "3007"
        LOG_LEVEL: "debug"
        NODE_ENV: "development"
        PORT: "3006"
        WAIT_SECONDS: "5"
        EOF
        destination = "local/file.env"
        env = true
      }
    }

    task "backend" {
      driver = "docker"

      config {
        image = "localhost:5002/rafiki-backend"
      }

      template {
        data = <<EOF
        DATABASE_URL="postgres://{{with secret "database-dev/creds/rafiki_backend"}}{{.Data.data.username}}:{{.Data.password}}{{end}}@localhost:5432/rafiki_auth?sslmode=disable"
        {{with secret "kv/data/dev/rafiki_backend/config"}}
        SIGNATURE_SECRET={{.Data.data.SIGNATURE_SECRET}}
        STREAM_SECRET={{.Data.data.STREAM_SECRET}}
        {{end}}
        EOF
        destination - "secrets/file.env"
        env = true
      }

      template {
        data = <<EOF
        ADMIN_PORT: "3001"
        AUTH_SERVER_GRANT_URL: "http://127.0.0.1:3006"
        AUTH_SERVER_INTROSPECTION_URL: "http://127.0.0.1:3007"
        CONNECTOR_PORT: "3002"
        EXCHANGE_RATES_LIFETIME: "15000"
        GRAPHQL_IDEMPOTENCY_KEY_TTL_MS: "8.64e+07"
        GRAPHQL_IDEMPOTENCY_KEY_LOCK_MS: "2000"
        ILP_ADDRESS: "test.cloud-nine-wallet"
        INCOMING_PAYMENT_WORKERS: "1"
        INCOMING_PAYMENT_WORKER_IDLE: "200"
        KEY_ID: "rafiki"
        LOG_LEVEL: "debug"
        NODE_ENV: "development"
        OPEN_PAYMENTS_PORT: "80"
        OPEN_PAYMENTS_URL: "http://127.0.0.1"
        OUTGOING_PAYMENT_WORKERS: "4"
        OUTGOING_PAYMENT_WORKER_IDLE: "200"
        PAYMENT_POINTER_URL: "https://rafiki.test/.well-known/pay"
        PAYMENT_POINTER_WORKERS: "1"
        PAYMENT_POINTER_WORKER_IDLE: "200"
        PRIVATE_KEY_FILE: ""
        PUBLIC_HOST: "https://rafiki.test"
        QUOTE_LIFESPAN: "300000"
        REDIS_TLS_CA_FILE_PATH: ""
        REDIS_TLS_CERT_FILE_PATH: ""
        REDIS_TLS_KEY_FILE_PATH: ""
        REDIS_URL: "redis://127.0.0.1:6379/1"
        SIGNATURE_VERSION: "1"
        SLIPPAGE: "0.01"
        TIGERBEETLE_CLUSTER_ID: none
        TIGERBEETLE_REPLICA_ADDRESSES: none
        USE_TIGERBEETLE: "false"
        WEBHOOK_TIMEOUT:  "200"
        WEBHOOK_URL: "http://127.0.0.1:8080/rafiki"
        WEBHOOK_WORKERS: "1"
        WEBHOOK_WORKER_IDLE: "200"
        WITHDRAWAL_THROTTLE_DELAY: ""
        EOF
        destination = "local/file.env"
        env = true
      }
    }

    task "frontend" {
      driver = "docker"

      config {
        image = "localhost:5002/rafiki-frontend"
      }

      template {
        data = <<EOF
        GRAPHQL_URL: "http://127.0.0.1:3001/graphql"
        LOG_LEVEL: debug
        NODE_ENV: development
        PORT: "3010"
        EOF
        destination = "local/file.env"
        env = true
      }
    }
  }

}
