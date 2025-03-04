job "protea" {
  datacenters = ["dc1"]
  type = "service"

  
  group "protea" {
    count = 1

    volume "protea" {
      type = "host"
      read_only = false
      source = "protea"
    }

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
      name = "protea"
      port = "http"
      tags = [
        "traefik.enable=true",
        "traefik.http.routers.protea.rule=Host(`wallet.interledger.test`)"
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
      name = "protea-ws"
      port = "websocket"
      tags = [
        "traefik.enable=true",
        "traefik.http.routers.protea-ws.rule=Host(`wallet.interledger.test`) && PathPrefix(`/socket`)"
      ]
    }

    task "protea" {
      driver = "docker"

      config {
        image = "localhost:5002/protea"
        volumes = ["/home/vagrant/fynbos/typescript/protea:/app"]
      }

      env {
        CHOKIDAR_USEPOLLING = true
        COOKIE_SECRETS = "[\"localsecret\"]"
        PUSHER_APP_KEY = "91988d6075551d29760a"
        PUSHER_APP_CLUSTER = "eu"
        KRATOS_URL = "http://127.0.0.1:4433"
        PAYMENT_POINTER_BASE = "local.fynbos.me"
        RAFIKI_AUTH_ENDPOINT = "http://127.0.0.1:3006"
        DATO_API_TOKEN = "b96709bce873f5280722da965b0e9d"
        FYNBOS_ENV = "local"
      }

      resources {
        memory = 1024
      }
    }
  }
}
