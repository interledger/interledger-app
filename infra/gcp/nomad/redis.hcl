job "redis" {
  datacenters = ["dc1"]
  type        = "service"
  namespace   = "dev"

  group "redis" {
    count = 1

    update {
      max_parallel     = 1
      canary           = 1
      auto_revert      = true
      auto_promote     = true
      health_check     = "checks"
    }

    ephemeral_disk {
      size = 300
    }

    network {
      mode = "bridge"
    }

    service {
      name = "dev-redis"
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
        memory = 1024
      }

    }
  }
}
