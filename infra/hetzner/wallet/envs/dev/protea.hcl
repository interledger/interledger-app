job "protea" {
  datacenters = ["dc1"]
  type = "service"
  namespace = "dev"

  
  group "protea" {
    count = 1

    network {
      mode = "bridge"
      port "http" {
        to = 3000
      }
    }

    service {
      connect {
        sidecar_service {
          proxy {
            upstreams {
              destination_name = "backend-grpc"
              local_bind_port  = 8443
            }
            upstreams {
              destination_name = "backend-admin"
              local_bind_port = 8448
            }
            upstreams {
              destination_name = "redis"
              local_bind_port = 6379
            }
            upstreams {
              destination_name = "kratos"
              local_bind_port = 4433
            }
          }
        }
      }
    }

    service {
      name = "protea"
      port = "http"
      tags = [
        "traefik.enable=true",
        "traefik.http.routers.protea.rule=Host(`eu1.fynbos.dev`)"
      ]
    }

    task "protea" {
      driver = "docker"

      config {
        image = "localhost:5002/protea"
      }

      template {
        data = <<EOF
        {{- with secret "kv/data/dev/protea/config" -}}
        GOOGLE_MAPS_API_KEY={{.Data.data.GOOGLE_MAPS_API_KEY}}
        COOKIE_SECRETS={{.Data.data.COOKIE_SECRETS}}
        DATO_API_TOKEN={{.Data.data.DATO_API_TOKEN}}
        BT_TOKEN={{.Data.data.BT_TOKEN}}
        SENTRY_DSN={{.Data.data.SENTRY_DSN}}
        SEGMENT_API_KEY={{.Data.data.SEGMENT_API_KEY}}
        CF_TURNSTILE_SITE_KEY={{.Data.data.CF_TURNSTILE_SITE_KEY}}
        CF_TURNSTILE_SECRET_KEY={{.Data.data.CF_TURNSTILE_SECRET_KEY}}
        {{- end -}}

        OTEL_EXPORTER_OTLP_ENDPOINT=grpc://api.honeycomb.io:443
        OTEL_EXPORTER_OTLP_HEADERS=x-honeycomb-team=oTtj00yo3Le8WofuInVwFB
        OTEL_SERVICE_NAME=protea
        GATE_SIGNUP=false
        PUSHER_APP_KEY=dee1c9a2f290e802ac26
        PUSHER_APP_CLUSTER=us2
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
