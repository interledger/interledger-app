job "postgres" {
  datacenters = ["dc1"]
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
        static = 7432
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

      template {
        data = <<EOH
          POSTGRES_USER=postgres
          POSTGRES_PASSWORD=password
        EOH

        destination = "secrets/file.env"
        env         = true
      }

      template {
        data = <<EOH
          CREATE DATABASE kratos;
          CREATE DATABASE backend;
          CREATE DATABASE pacioli;
          CREATE DATABASE temporal;
          CREATE DATABASE temporal_visibility;
          CREATE DATABASE rafiki_backend;
          CREATE DATABASE rafiki_auth;
          CREATE DATABASE mockbos;
          \c backend;
          CREATE EXTENSION pg_trgm;
        EOH

        destination = "local/init-user-db.sql"
        env = false
      }

      config {
        image = "postgres:16.1"

        volumes = [
          "local/init-user-db.sql:/docker-entrypoint-initdb.d/init-user-db.sql",
          "/opt/nomad/data/volume/postgres:/var/lib/postgresql/data"
        ]
      }

      resources {
        cpu = 500
      }
    }
  }
}
