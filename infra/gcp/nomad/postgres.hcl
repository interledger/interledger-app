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
      name = "dev-postgres"
      port = 5432
      connect {
        sidecar_service {}
      }
    }

    vault {}

    task "postgres" {
      driver = "docker"

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
          -- Create Kratos user and DB
          CREATE USER kratos;
          CREATE DATABASE kratos;
          GRANT ALL ON DATABASE kratos to kratos;
          GRANT USAGE ON ALL SEQUENCES IN SCHEMA public TO kratos;
          GRANT ALL ON SCHEMA public TO kratos;
          ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT SELECT, INSERT, UPDATE, DELETE ON TABLES TO kratos;

          -- Create backend user and DB
          CREATE DATABASE backend;
          CREATE USER backend;
          GRANT ALL ON DATABASE backend to backend;
          GRANT USAGE ON ALL SEQUENCES IN SCHEMA public TO backend;
          GRANT ALL ON SCHEMA public TO backend;
          ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT SELECT, INSERT, UPDATE, DELETE ON TABLES TO backend;

          -- Create pacioli user and DB
          CREATE DATABASE  pacioli;          
          CREATE USER pacioli;
          GRANT ALL ON DATABASE pacioli to pacioli;
          GRANT USAGE ON ALL SEQUENCES IN SCHEMA public TO pacioli;
          GRANT ALL ON SCHEMA public TO pacioli;
          ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT SELECT, INSERT, UPDATE, DELETE ON TABLES TO pacioli;
          
          -- Create temporal user and DB
          CREATE DATABASE temporal;
          CREATE USER temporal;
          GRANT ALL ON DATABASE temporal to temporal;
          GRANT USAGE ON ALL SEQUENCES IN SCHEMA public TO temporal;
          GRANT ALL ON SCHEMA public TO temporal;
          ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT SELECT, INSERT, UPDATE, DELETE ON TABLES TO temporal;

          CREATE DATABASE temporal_visibility;
          CREATE USER temporal_visibility;
          GRANT ALL ON DATABASE temporal_visibility to temporal_visibility;
          GRANT USAGE ON ALL SEQUENCES IN SCHEMA public TO temporal_visibility;
          GRANT ALL ON SCHEMA public TO temporal_visibility;
          ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT SELECT, INSERT, UPDATE, DELETE ON TABLES TO temporal_visibility;

          -- Create rafiki user and DB
          CREATE DATABASE rafiki_backend;
          CREATE USER rafiki_backend;
          GRANT ALL ON DATABASE rafiki_backend to rafiki_backend;
          GRANT USAGE ON ALL SEQUENCES IN SCHEMA public TO rafiki_backend;
          GRANT ALL ON SCHEMA public TO rafiki_backend;
          ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT SELECT, INSERT, UPDATE, DELETE ON TABLES TO rafiki_backend;

          CREATE DATABASE rafiki_auth;
          CREATE USER rafiki_auth;
          GRANT ALL ON DATABASE rafiki_auth to rafiki_auth;
          GRANT USAGE ON ALL SEQUENCES IN SCHEMA public TO rafiki_auth;
          GRANT ALL ON SCHEMA public TO rafiki_auth;
          ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT SELECT, INSERT, UPDATE, DELETE ON TABLES TO rafiki_auth;
        EOH

        destination = "$NOMAD_TASK_DIR/init-user-db.sql"
        env = false
      }

      config {
        image = "postgres:16.1"

        volumes = [
        "$NOMAD_TASK_DIR:/docker-entrypoint-initdb.d",
        "/data/live/nomad/data/volume/db-dev/data:/var/lib/postgresql/data"
        ]
      }

      resources {
        cpu = 500
        memory = 2048
      }
    }
  }
}
