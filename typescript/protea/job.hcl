job "protea" {
  datacenters = ["dc1"]

  type = "service"

  group "protea" {
    count = 1

    network {
      mode = "bridge"
      port "http" {
        to = 3000
      }
    }

    service {
      name = "protea"
      port = "http"

      tags = [
        "traefik.enable=true",
        "traefik.http.routers.protea.rule=Host(`fynbos.test`)",
        "traefik.http.routers.protea.entrypoints=http",
      ]

      connect {
        sidecar_service {
          proxy {
            upstreams {
              destination_name = "backend"
              local_bind_port = 8443
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

    task "protea" {

      driver = "docker"

      config {
        image = "protea:local"
        ports = ["http"]
      }

      resources {
        cpu    = 1000
        memory = 1028
      }

      env {
        DATO_API_TOKEN = "b96709bce873f5280722da965b0e9d"
        BACKEND_GRPC_URL = "http://localhost:8443"
        REDIS_URL = "redis://localhost:6379"
        KRATOS_URL = "http://localhost:4433"
      }
    }
  }
}