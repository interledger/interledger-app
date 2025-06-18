job "mockbos" {
  datacenters = ["dc1"]
  type = "service"
  
  group "mockbos" {
    count = 1

    network {
      mode = "bridge"
      port "http" {
        to = 8080
      }
    }

    service {
      name = "mockbos"
      port = 8080
      connect {
        sidecar_service {
          tags = [
            "traefik.enable = false"
          ]

          proxy {
            upstreams {
              destination_name = "postgres"
              local_bind_port  = 5432
            }
            upstreams {
              destination_name = "backend-http"
              local_bind_port = 9090
            }
          }
        }
      }
    }

    volume "go" {
      type = "host"
      read_only = false
      source = "go"
    }

    task "mockbos" {
      driver = "docker"

      config {
        image = "localhost:5002/mockbos"
        entrypoint = ["/go/bin/air"]
        args =  ["--build.poll", "true", "--build.include_ext", "go,hcl", "--build.cmd", "go build -o /local/main /build/mockbos/main.go", "--build.bin", "/local/main"]
        volumes = ["/home/vagrant/fynbos/go:/build"]
      }

      env {
          DB_URL = "postgres://postgres:password@127.0.0.1:5432/mockbos?sslmode=disable"
          XAGO_WEBHOOK_URL = "http://localhost:9090/webhooks/xago"
      }

      resources {
        memory_max = 750
      }
    }

  }
}
