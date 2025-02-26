job "rafiki" {
  datacenters = ["dc1"]
  type = "service"
  namespace = "dev"

  group "rafiki" {
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

      port "connector" {
        to = 3002
      }
    }

    service {
      name = "dev-rafiki"
      connect {
        sidecar_service {
          proxy {
            upstreams {
              destination_name = "dev-postgres"
              local_bind_port  = 5432
            }
            upstreams {
              destination_name = "dev-redis"
              local_bind_port = 6379
            }
            upstreams {
              destination_name = "dev-backend-http"
              local_bind_port = 8080
            }
          }
        }
      }
    }

    service {
      name = "dev-rafiki-backend-admin"
      port = 3001
      connect {
        sidecar_service {}
      }
    }

    service {
      name = "dev-rafiki-auth"
      port = 3009
      connect {
        sidecar_service {}
      }
    }

    service {
      name = "dev-rafiki-connector"
      port = "connector"
      tags = [
        "traefik.enable=true",
        "traefik.http.middlewares.dev-rafiki-connector-stripprefix.stripprefix.prefixes=/ilp",
        "traefik.http.routers.dev-rafiki-connector.rule=Host(`eu1.fynbos.me`) && Path(`/ilp`)",
        "traefik.http.routers.dev-rafiki-connector.middlewares=dev-rafiki-connector-stripprefix@consulcatalog"
      ]
    }

    service {
      name = "dev-rafiki-openpayments"
      port = "open-payments"
      tags = [
        "traefik.enable=true",
        "traefik.http.routers.dev-rafiki-incomingpayments.rule=Host(`eu1.fynbos.me`) && PathPrefix(`/incoming-payments`)",
        "traefik.http.routers.dev-rafiki-outgoingpayments.rule=Host(`eu1.fynbos.me`) && PathPrefix(`/outgoing-payments`)",
        "traefik.http.routers.dev-rafiki-quotes.rule=Host(`eu1.fynbos.me`) && PathPrefix(`/quotes`)",
        "traefik.http.routers.dev-rafiki-wellknown.rule=Host(`eu1.fynbos.me`) && PathPrefix(`/.well-known`)"
      ]

      check {    
        type     = "http"
        port     = "open-payments"
        path     = "/healthz"
        interval = "5s"
        timeout  = "2s"
      }
    }

    service {
      name = "dev-rafiki-auth-public"
      port = "auth"
      tags = [
        "traefik.enable=true",
        "traefik.http.middlewares.dev-rafiki-auth-stripprefix.stripprefix.prefixes=/gnap",
        "traefik.http.routers.dev-rafiki-auth.rule=((Host(`eu1.fynbos.me`) && Path(`/gnap`)) || Host(`auth.eu1.fynbos.dev`))",
        "traefik.http.routers.dev-rafiki-auth-token.rule=Host(`eu1.fynbos.me`) && PathPrefix(`/token`)",
        "traefik.http.routers.dev-rafiki-auth-interact.rule=Host(`eu1.fynbos.me`) && PathPrefix(`/interact`)",
        "traefik.http.routers.dev-rafiki-auth-continue.rule=Host(`eu1.fynbos.me`) && PathPrefix(`/continue`)",
        "traefik.http.routers.dev-rafiki-auth.middlewares=dev-rafiki-auth-stripprefix@consulcatalog"
      ]

      check {    
        type     = "http"
        port     = "auth"
        path     = "/healthz"
        interval = "5s"
        timeout  = "2s"
      }
    }

    service {
      name = "dev-rafiki-auth-admin"
      port = 3003
      connect {
        sidecar_service {}
      }
    }

    service {
      name = "dev-rafiki-frontend"
      port = "frontend"
      tags = [
        "traefik.enable=true",
        "traefik.http.routers.dev-rafiki-frontend.rule=Host(`rafiki-dev.mgnt.fynbos.dev`)"
      ]
    }

    task "auth" {
      driver = "docker"
      kill_timeout = "120s"

      config {
        image = format("registry.gitlab.com/fynbos/fynbos/rafiki-auth:%s", "v1.0.0-alpha.15")
      }

      vault {}

      template {
        data = <<EOF
        {{with secret "database-dev/static-creds/rafiki_auth"}}
        AUTH_DATABASE_URL="postgres://{{.Data.username}}:{{.Data.password}}@127.0.0.1:5432/rafiki_auth?sslmode=disable"
        {{end}}

        {{with secret "kv/data/dev/rafiki/config"}}
        IDENTITY_SERVER_SECRET={{.Data.data.IDENTITY_SERVER_SECRET}}
        COOKIE_KEY={{.Data.data.COOKIE_KEY}}
        {{end}}
        EOF
        destination = "secrets/file.env"
        env = true
      }

      template {
        data = <<EOF
        ACCESS_TOKEN_DELETION_DAYS="30"
        ACCESS_TOKEN_EXPIRY_SECONDS="600"
        ADMIN_PORT="3003"
        AUTH_PORT="3006"
        AUTH_SERVER_URL="https://auth.eu1.fynbos.dev"
        DATABASE_CLEANUP_WORKERS="1"
        ENV_FILE=
        IDENTITY_SERVER_URL="https://eu1.fynbos.dev/consent"
        INCOMING_PAYMENT_INTERACTION="false"
        QUOTE_INTERACTION="false"
        INTROSPECTION_PORT="3007"
        LOG_LEVEL="debug"
        NODE_ENV="production"
        PORT="3006"
        WAIT_SECONDS="5"
        TRUST_PROXY="true"
        REDIS_TLS_CA_FILE_PATH=""
        REDIS_TLS_CERT_FILE_PATH=""
        REDIS_TLS_KEY_FILE_PATH=""
        REDIS_URL="redis://127.0.0.1:6379/2"
        EOF
        destination = "local/file.env"
        env = true
      }
    }

    task "backend" {
      driver = "docker"
      kill_timeout = "120s"

      config {
        image = format("registry.gitlab.com/fynbos/fynbos/rafiki-backend:%s", "v1.0.0-alpha.15")
        privileged = true
      }

      vault {}

      template {
        data = <<EOF
        {{with secret "database-dev/static-creds/rafiki_backend"}}
        DATABASE_URL="postgres://{{.Data.username}}:{{.Data.password}}@127.0.0.1:5432/rafiki_backend?sslmode=disable"
        {{end}}

        {{with secret "kv/data/dev/rafiki/config"}}
        SIGNATURE_SECRET={{.Data.data.SIGNATURE_SECRET}}
        STREAM_SECRET={{.Data.data.STREAM_SECRET}}
        {{end}}
        EOF
        destination = "secrets/file.env"
        env = true
      }

      template {
        data = <<EOF
        ADMIN_PORT="3001"
        AUTH_SERVER_GRANT_URL="http://127.0.0.1:3006"
        AUTH_SERVER_INTROSPECTION_URL="http://127.0.0.1:3007"
        CONNECTOR_PORT="3002"
        ENV_FILE=
        EXCHANGE_RATES_LIFETIME="15000"
        GRAPHQL_IDEMPOTENCY_KEY_TTL_MS="8.64e+07"
        GRAPHQL_IDEMPOTENCY_KEY_LOCK_MS="2000"
        ILP_ADDRESS="test.fynbos"
        ILP_CONNECTOR_URL="https://eu1.fynbos.me/ilp"
        INCOMING_PAYMENT_WORKERS="1"
        INCOMING_PAYMENT_WORKER_IDLE="200"
        INSTANCE_NAME="fynbosdev"
        KEY_ID="rafikidev"
        LOG_LEVEL="debug"
        NODE_ENV="production"
        OPEN_PAYMENTS_PORT="80"
        OPEN_PAYMENTS_URL="https://eu1.fynbos.me"
        OUTGOING_PAYMENT_WORKERS="4"
        OUTGOING_PAYMENT_WORKER_IDLE="200"
        PAYMENT_POINTER_URL="https://eu1.fynbos.me/.well-known/pay"
        PAYMENT_POINTER_WORKERS="1"
        PAYMENT_POINTER_WORKER_IDLE="200"
        PRIVATE_KEY_FILE=""
        PUBLIC_HOST="https://eu1.fynbos.me"
        QUOTE_LIFESPAN="300000"
        REDIS_TLS_CA_FILE_PATH=""
        REDIS_TLS_CERT_FILE_PATH=""
        REDIS_TLS_KEY_FILE_PATH=""
        REDIS_URL="redis://127.0.0.1:6379/1"
        SIGNATURE_VERSION="1"
        SLIPPAGE="0.01"
        TIGERBEETLE_CLUSTER_ID=none
        TIGERBEETLE_REPLICA_ADDRESSES=none
        TRUST_PROXY="true"
        USE_TIGERBEETLE=false
        WEBHOOK_TIMEOUT= "200"
        WEBHOOK_URL="http://127.0.0.1:8080/rafiki"
        WEBHOOK_WORKERS="1"
        WEBHOOK_WORKER_IDLE="200"
        WITHDRAWAL_THROTTLE_DELAY=""
        EOF
        destination = "local/file.env"
        env = true
      }
    }

    task "frontend" {
      driver = "docker"

      config {
        image = format("registry.gitlab.com/fynbos/fynbos/rafiki-frontend:%s", "v1.0.0-alpha.16")
      }

      template {
        data = <<EOF
        GRAPHQL_URL="http://127.0.0.1:3001/graphql"
        LOG_LEVEL=debug
        NODE_ENV=production
        PORT="3010"
        AUTH_ENABLED=false
        OPEN_PAYMENTS_URL="https://eu1.fynbos.me"
        EOF
        destination = "local/file.env"
        env = true
      }
    }
  }

}
