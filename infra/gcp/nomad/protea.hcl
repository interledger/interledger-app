variable "image_hash" {
  type = string
}

job "protea" {
  datacenters = ["dc1"]
  type = "service"
  namespace = "dev"

  
  group "protea" {
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
        to = 3000
      }
    }

    service {
      name = "dev-protea"
      port = "http"
      tags = [
        "traefik.enable=true",
        "traefik.http.routers.dev-protea.rule=Host(`eu1.fynbos.dev`)"
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
              destination_name = "dev-backend-grpc"
              local_bind_port  = 8443
            }
            upstreams {
              destination_name = "dev-backend-admin"
              local_bind_port = 8448
            }
            upstreams {
              destination_name = "dev-redis"
              local_bind_port = 6379
            }
            upstreams {
              destination_name = "dev-kratos"
              local_bind_port = 4433
            }
            upstreams {
              destination_name = "dev-rafiki-auth"
              local_bind_port = 3009
            }
          }
        }
      }
    }

    task "protea" {
      driver = "docker"

      config {
        image = format("registry.gitlab.com/fynbos/fynbos/protea:%s", var.image_hash)
      }

      vault {}

      template {
        data = <<EOF
        {{ with secret "kv/data/dev/protea/config" }}
        GOOGLE_MAPS_API_KEY='{{.Data.data.GOOGLE_MAPS_API_KEY}}'
        COOKIE_SECRETS='{{.Data.data.COOKIE_SECRETS}}'
        DATO_API_TOKEN='{{.Data.data.DATO_API_TOKEN}}'
        BT_TOKEN='{{.Data.data.BT_TOKEN}}'
        SENTRY_DSN='{{.Data.data.SENTRY_DSN}}'
        SEGMENT_API_KEY='{{.Data.data.SEGMENT_API_KEY}}'
        CF_TURNSTILE_SITE_KEY='{{.Data.data.CF_TURNSTILE_SITE_KEY}}'
        CF_TURNSTILE_SECRET_KEY='{{.Data.data.CF_TURNSTILE_SECRET_KEY}}'
        RAFIKI_AUTH_SECRET='{{.Data.data.RAFIKI_AUTH_SECRET}}'
        {{ end }}

        OTEL_EXPORTER_OTLP_ENDPOINT=grpc://api.honeycomb.io:443
        OTEL_EXPORTER_OTLP_HEADERS=x-honeycomb-team=oTtj00yo3Le8WofuInVwFB
        OTEL_SERVICE_NAME=protea
        GATE_SIGNUP=false
        PUSHER_APP_KEY=dee1c9a2f290e802ac26
        PUSHER_APP_CLUSTER=us2
        KRATOS_URL=http://127.0.0.1:4433
        PAYMENT_POINTER_BASE=eu1.fynbos.me
        RAFIKI_AUTH_ENDPOINT=http://127.0.0.1:3009
        FYNBOS_ENV=dev
        EOF
        destination = "secrets/file.env"
        env         = true
      }

      resources {
        memory = 1024
      }
    }
  }
}
