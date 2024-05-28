job "botanist" {
  datacenters = ["dc1"]
  type = "service"
  namespace = "dev"
  
  group "botanist" {
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
            upstreams {
              destination_name = "backend-admin"
              local_bind_port = 8448
            }
          }
        }
      }
    }

    service {
      name = "botanist"
      port = "http"
      tags = [
        "traefik.enable=true",
        "traefik.http.routers.botanist.rule=Host(`admin-dev.mgnt.fynbos.dev`)"
      ]
    }

    task "botanist" {
      driver = "docker"

      config {
        image = "localhost:5002/botanist"
      }

      template {
        data = <<EOF
        BACKEND_GRPC_URL=0.0.0.0:8448
        FYNBOS_ENV=dev
        EOF
        destination = "local/file.env"
        env         = true
      }

      resources {
        memory = 1024
      }
    }
  }
}
