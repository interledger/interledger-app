job "rafiki" {
  datacenters = ["dc1"]
  type = "service"

  group "rafiki" {
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

      port "connector" {
        to = 3002
      }
    }

    service {
      name = "rafiki"
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
      name = "rafiki-auth"
      port = 3006
      connect {
        sidecar_service {}
      }
    }

    service {
      name = "rafiki-connector"
      port = "connector"
      tags = [
        "traefik.enable=true",
        "traefik.http.middlewares.rafiki-connector-stripprefix.stripprefix.prefixes=/ilp",
        "traefik.http.routers.rafiki-connector.rule=Host(`local.fynbos.me`) && Path(`/ilp`)",
        "traefik.http.routers.rafiki-connector.middlewares=rafiki-connector-stripprefix@consulcatalog"
      ]
    }

    service {
      name = "rafiki-openpayments"
      port = "open-payments"
      tags = [
        "traefik.enable=true",
        "traefik.http.routers.rafiki-incomingpayments.rule=Host(`local.fynbos.me`) && PathPrefix(`/incoming-payments`)",
        "traefik.http.routers.rafiki-outgoingpayments.rule=Host(`local.fynbos.me`) && PathPrefix(`/outgoing-payments`)",
        "traefik.http.routers.rafiki-quotes.rule=Host(`local.fynbos.me`) && PathPrefix(`/quotes`)",
        "traefik.http.routers.rafiki-wellknown.rule=Host(`local.fynbos.me`) && PathPrefix(`/.well-known`)"
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
      name = "rafiki-auth-public"
      port = "auth"
      tags = [
        "traefik.enable=true",
        "traefik.http.middlewares.rafiki-auth-stripprefix.stripprefix.prefixes=/gnap",
        "traefik.http.routers.rafiki-auth.rule=((Host(`local.fynbos.me`) && Path(`/gnap`)) || Host(`auth.fynbos.test`))",
        "traefik.http.routers.rafiki-auth-token.rule=Host(`local.fynbos.me`) && PathPrefix(`/token`)",
        "traefik.http.routers.rafiki-auth-interact.rule=Host(`local.fynbos.me`) && PathPrefix(`/interact`)",
        "traefik.http.routers.rafiki-auth-continue.rule=Host(`local.fynbos.me`) && PathPrefix(`/continue`)",
        "traefik.http.routers.rafiki-auth.middlewares=rafiki-auth-stripprefix@consulcatalog"
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
        "traefik.http.routers.rafiki-frontend.rule=Host(`rafiki.mgnt.fynbos.test`)"
      ]
    }

    task "auth" {
      driver = "docker"

      config {
        image = "localhost:5002/rafiki-auth"
      }

      env {
        AUTH_DATABASE_URL = "postgres://postgres:password@127.0.0.1:5432/rafiki_auth?sslmode=disable"
        IDENTITY_SERVER_SECRET = "changeme"
        COOKIE_KEY = ""
        ACCESS_TOKEN_DELETION_DAYS = "30"
        ACCESS_TOKEN_EXPIRY_SECONDS = "600"
        ADMIN_PORT = "3003"
        AUTH_PORT = "3006"
        AUTH_SERVER_DOMAIN = "https://auth.fynbos.test"
        DATABASE_CLEANUP_WORKERS = "1"
        ENV_FILE = "" 
        IDENTITY_SERVER_DOMAIN = "https://wallet.fynbos.test/consent"
        INCOMING_PAYMENT_INTERACTION = "false"
        QUOTE_INTERACTION = "false"
        INTROSPECTION_PORT = "3007"
        LOG_LEVEL = "debug"
        NODE_ENV = "production"
        PORT = "3006"
        WAIT_SECONDS = "5"
        TRUST_PROXY = "true"
        REDIS_TLS_CA_FILE_PATH = ""
        REDIS_TLS_CERT_FILE_PATH = ""
        REDIS_TLS_KEY_FILE_PATH = ""
        REDIS_URL = "redis://127.0.0.1:6379/2"
      }
    }

    task "backend" {
      driver = "docker"

      config {
        image = "localhost:5002/rafiki-backend"
        privileged = true
      }

      env {
        DATABASE_URL = "postgres://postgres:password@127.0.0.1:5432/rafiki_backend?sslmode=disable"
        SIGNATURE_SECRET = "overridethisValue"
        STREAM_SECRET = "BjPXtnd00G2mRQwP/8ZpwyZASOch5sUXT5o0iR5b5wU="
        ADMIN_PORT = "3001"
        AUTH_SERVER_GRANT_URL = "http://127.0.0.1:3006"
        AUTH_SERVER_INTROSPECTION_URL = "http://127.0.0.1:3007"
        CONNECTOR_PORT = "3002"
        ENV_FILE = ""
        EXCHANGE_RATES_LIFETIME = "15000"
        GRAPHQL_IDEMPOTENCY_KEY_TTL_MS = "8.64e+07"
        GRAPHQL_IDEMPOTENCY_KEY_LOCK_MS = "2000"
        IDENTITY_SERVER_DOMAIN = "https://wallet.fynbos.test/consent"
        ILP_ADDRESS = "test.fynbos"
        ILP_CONNECTOR_ADDRESS = "https://local.fynbos.me/ilp"
        INCOMING_PAYMENT_WORKERS = "1"
        INCOMING_PAYMENT_WORKER_IDLE = "200"
        INSTANCE_NAME = "fynbosdev"
        KEY_ID = "rafikidev"
        LOG_LEVEL = "debug"
        NODE_ENV = "production"
        OPEN_PAYMENTS_PORT = "80"
        OPEN_PAYMENTS_URL = "https://local.fynbos.me"
        OUTGOING_PAYMENT_WORKERS = "4"
        OUTGOING_PAYMENT_WORKER_IDLE = "200"
        PAYMENT_POINTER_URL = "https://local.fynbos.me/.well-known/pay"
        PAYMENT_POINTER_WORKERS = "1"
        PAYMENT_POINTER_WORKER_IDLE = "200"
        PRIVATE_KEY_FILE = ""
        PUBLIC_HOST = "https://local.fynbos.me"
        QUOTE_LIFESPAN = "300000"
        REDIS_TLS_CA_FILE_PATH = ""
        REDIS_TLS_CERT_FILE_PATH = ""
        REDIS_TLS_KEY_FILE_PATH = ""
        REDIS_URL = "redis://127.0.0.1:6379/1"
        SIGNATURE_VERSION = "1"
        SLIPPAGE = "0.01"
        TIGERBEETLE_CLUSTER_ID = "none"
        TIGERBEETLE_REPLICA_ADDRESSES = "none"
        TRUST_PROXY = "true"
        USE_TIGERBEETLE = "false"
        WEBHOOK_TIMEOUT= "200"
        WEBHOOK_URL = "http://127.0.0.1:8080/rafiki"
        WEBHOOK_WORKERS = "1"
        WEBHOOK_WORKER_IDLE = "200"
        WITHDRAWAL_THROTTLE_DELAY = ""
      }
    }

    task "frontend" {
      driver = "docker"

      config {
        image = "localhost:5002/rafiki-frontend"
      }

      env {
        GRAPHQL_URL = "http://127.0.0.1:3001/graphql"
        LOG_LEVEL = "debug"
        NODE_ENV = "production"
        PORT = "3010"
      }
    }
  }

}
