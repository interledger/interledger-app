job "postgres" {
  datacenters = [
    "dc1"]
  type = "service"

  group "postgres" {
    count = 1

    volume "postgres" {
      type = "host"
      read_only = false
      source = "postgres"
    }

    restart {
      attempts = 10
      interval = "5m"
      delay = "25s"
      mode = "delay"
    }

    network {
      mode = "bridge"
      port "postgres" {
        to = 5432
      }
    }

    service {
      name = "postgres"
      port = 5432
      connect {
        sidecar_service {}
      }
    }

    task "postgres" {
      driver = "docker"

      volume_mount {
        volume = "postgres"
        destination = "/var/lib/postgresql"
        read_only = false
      }

      template {
        data = <<EOH
          POSTGRES_USER=postgres
          POSTGRES_PASSWORD=password
        EOH

        destination = "secrets/file.env"
        env         = true
      }

      config {
        image = "postgres:16.1"
      }

      resources {
        cpu = 500
        memory = 2048
      }
    }
  }
}
