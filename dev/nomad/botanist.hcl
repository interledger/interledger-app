job "botanist" {
  datacenters = ["dc1"]
  type = "service"
  
  group "botanist" {
    count = 1

    network {
      mode = "bridge"
      port "http" {
        to = 3000
      }
    }

    volume "botanist" {
      type = "host"
      read_only = false
      source = "botanist"
    }

    service {
      name = "botanist"
      port = "http"
      tags = [
        "traefik.enable=true",
        "traefik.http.routers.dev-botanist.rule=Host(`admin.mgnt.fynbos.test`)"
      ]

      check {    
        type     = "http"
        port     = "http"
        path     = "/"
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
              destination_name = "backend"
              local_bind_port = 8448
            }
          }
        }
      }
    }

    task "botanist" {
      driver = "docker"

      config {
        image = "http://localhost:5002/botanist"
        volumes = ["/home/vagrant/fynbos/typescript/botanist:/app"]
      }

      env {
        CHOKIDAR_USEPOLLING = "true"
        BACKEND_GRPC_URL = "0.0.0.0:8448"
        FYNBOS_ENV = "dev"
      }

      resources {
        memory = 1024
      }
    }
  }
}
