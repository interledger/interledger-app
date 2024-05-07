job "postgres" {
  datacenters = ["dc1"]
  type = "service"
  namespace = "dev"

  group "postgres" {
    count = 1

    volume "postgres" {
      type = "host"
      read_only = false
      source = "db-dev"
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
        static = 6432
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

      vault {}

      template {
        data = <<EOH
          POSTGRES_USER=postgres
          POSTGRES_PASSWORD={{with secret "kv/data/dev/postgres/config"}}{{.Data.data.password}}{{end}}
        EOH

        destination = "secrets/file.env"
        env         = true
      }

      template {
        data = <<EOH
          -- Create an ADMIN user
          CREATE ROLE roach WITH LOGIN SUPERUSER PASSWORD 'roach';
          CREATE ROLE vault WITH LOGIN SUPERUSER PASSWORD 'vault';

          -- Create Kratos user and DB
          CREATE DATABASE kratos;

          -- Create backend user and DB
          CREATE DATABASE backend;

          -- Create pacioli user and DB
          CREATE DATABASE  pacioli;
          
          -- Create temporal user and DB
          CREATE DATABASE temporal;
          CREATE DATABASE temporal_visibility;

          -- Create rafiki user and DB
          CREATE DATABASE rafiki_backend;
          CREATE DATABASE rafiki_auth;
        EOH

        destination = "$NOMAD_TASK_DIR/init-user-db.sql"
        env = false
      }

      config {
        image = "postgres:16.1"

        volumes = ["$NOMAD_TASK_DIR:/docker-entrypoint-initdb.d"]
      }

      resources {
        cpu = 500
      }
    }
  }
}
