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

      template {
        data = <<EOH
          -- Create an ADMIN user
          CREATE ROLE roach WITH LOGIN SUPERUSER PASSWORD 'roach';

          -- Create Kratos user and DB
          CREATE DATABASE kratos;
          CREATE USER kratos;
          GRANT ALL ON DATABASE kratos TO kratos;

          -- Create backend user and DB
          CREATE DATABASE backend;
          CREATE USER backend;
          GRANT ALL ON DATABASE backend TO backend;

          -- Create pacioli user and DB
          CREATE DATABASE  pacioli;
          CREATE USER pacioli;
          GRANT ALL ON DATABASE pacioli TO backend;
          
          -- Create temporal user and DB
          CREATE USER temporal;
          CREATE DATABASE temporal;
          GRANT ALL ON DATABASE temporal TO temporal;
          CREATE DATABASE temporal_visibility;
          GRANT ALL ON DATABASE temporal_visibility TO temporal;
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
