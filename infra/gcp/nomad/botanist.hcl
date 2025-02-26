variable "image_hash" {
  type = string
}

job "botanist" {
  datacenters = ["dc1"]
  type = "service"
  namespace = "dev"
  
  group "botanist" {
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
      name = "dev-botanist"
      port = "http"
      tags = [
        "traefik.enable=true",
        "traefik.http.routers.dev-botanist.rule=Host(`admin-dev.mgnt.fynbos.dev`)"
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
              destination_name = "dev-backend-admin"
              local_bind_port = 8448
            }
          }
        }
      }
    }

    task "botanist" {
      driver = "docker"

      config {
        image = format("registry.gitlab.com/fynbos/fynbos/botanist:%s", var.image_hash)
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
