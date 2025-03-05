job "temporal" {
  datacenters = [
    "dc1"
  ]
  type = "service"

  group "temporal" {
    count = 1

    network {
      mode = "bridge"
      port "temporal" {
        to = 7233
      }

      port "ui" {
        to = 8233
      }
    }

    service {
      name = "temporalui"
      port = "ui"
      tags = [
        "traefik.enable=true",
        "traefik.http.routers.temporalui.rule=Host(`temporal.mgnt.interledger.test`)"
      ]
    }

    service {
      name = "temporal"
      port = 7233
      connect {
        sidecar_service {}
      }
    }

    task "temporal" {
      driver = "docker"
      config {
        image = "localhost:5002/temporal"
      }

      resources {
        memory = 1024
      }
    }
  }
}
