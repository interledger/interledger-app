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
      port "websocket" {
        to = 8002
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
        "traefik.http.routers.protea.rule=Host(`fynbos.test`)"
      ]
    }

    service {
      name = "protea-ws"
      port = "websocket"
      tags = [
        "traefik.enable=true",
        "traefik.http.routers.protea-ws.rule=Host(`fynbos.test`) && PathPrefix(`/socket`)"
      ]
    }

    task "protea" {
      driver = "docker"

      config {
        image = "localhost:5002/protea"
      }

      template {
        data = file("protea.env")
        destination = "secrets/file.env"
        env         = true
      }

      resources {
        memory = 1024
      }
    }
  }
}
