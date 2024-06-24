job "redis" {
  datacenters = ["dc1"]
  type        = "service"


  group "redis" {
    count = 1

    ephemeral_disk {
      size = 300
    }

    network {
      mode = "bridge"
      port "redis" {
        to = 6379
      }
    }

    service {
      name = "redis"
      port = 6379
      connect {
        sidecar_service {}
      }
    }

    task "redis" {
      driver = "docker"

      config {
        image = "redis:alpine"
      }

      resources {
        cpu    = 500
        memory = 256
      }

    }
  }
}
