job "protea" {
  datacenters = ["dc1"]
  type = "service"

  
  group "protea" {
    count = 1

    network {
      mode = "bridge"
      port "http" {
        to = 8080
      }
    }

    service {
      connect {
        sidecar_service {
          proxy {
            upstreams {
              destination_name = "backend"
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
