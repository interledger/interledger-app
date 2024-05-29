job "kratos" {
  datacenters = [
    "dc1"
  ]
  type = "service"
  namespace = "dev"

  group "kratos" {
    count = 1

    network {
      mode = "bridge"
      port "http-public" {
        to = 4433
      }
      port "http-admin" {
        to = 4434
      }
    }

    service {
      connect {
        sidecar_service {
          proxy {
            upstreams {
              destination_name = "postgres"
              local_bind_port  = 5432
            }
          }
        }
      }
    }

    service {
      name = "kratos"
      port = 4433
      connect {
        sidecar_service {}
      }
    }

    service {
      name = "kratos-admin"
      port = 4434
      connect {
        sidecar_service {}
      }
    }

    task "kratos-migrate" {
      driver = "docker"

      config {
        image   = "oryd/kratos:v1.0.0"
        args    = [
          "migrate",
          "sql",
          "-e",
          "--yes",
          "--config",
          "/local/kratos.yaml"
        ]
      }

      template {
        data = file("kratos.yaml")
        destination = "local/kratos.yaml"
      }

      template {
        data = <<EOH
          DSN={{with secret "database-dev/creds/kratos"}}"postgres://{{.Data.username}}:{{.Data.password}}@localhost:5432/kratos?sslmode=disable"{{end}}
        EOH
        destination = "secrets/file.env"
        env = true
      }

      vault {}

      lifecycle {
        hook    = "prestart"
        sidecar = false
      }
    }

    task "kratos" {
      driver = "docker"

      template {
        data = file("kratos.yaml")
        destination = "local/kratos.yaml"
      }

      template {
        data = file("kratos-identity.json")
        destination = "local/identity.schema.json"
      }

      template {
        data = file("kratos-sendgrid.jsonnet")
        destination = "local/sendgrid.jsonnet"
      }

      template {
        data = <<EOH
          DSN={{with secret "database-dev/creds/kratos"}}"postgres://{{.Data.username}}:{{.Data.password}}@localhost:5432/kratos?sslmode=disable"{{end}}
          {{- with secret "kv/data/dev/kratos/config" -}}
          SECRETS_DEFAULT={{.Data.data.SECRETS_DEFAULT}}
          SECRETS_COOKIE={{.Data.data.SECRETS_COOKIE}}
          SECRETS_CIPHER={{.Data.data.SECRETS_CIPHER}}
          COURIER_HTTP_REQUEST_CONFIG_AUTH_CONFIG_VALUE="Bearer {{.Data.data.SENDGRID_API_KEY}}"
          {{- end -}}
        EOH
        destination = "secrets/file.env"
        env = true
      }

      vault {}

      config {
        image   = "oryd/kratos:v1.0.0"
        args    = [
          "serve",
          "all",
          "--config",
          "/local/kratos.yaml"
        ]
      }

      resources {
        cpu    = 100
        memory = 256
      }
    }

    task "kratos-mail" {
      driver = "docker"

      template {
        data = file("kratos.yaml")
        destination = "local/kratos.yaml"
      }

      template {
        data = file("kratos-identity.json")
        destination = "local/identity.schema.json"
      }

      template {
        data = file("kratos-sendgrid.jsonnet")
        destination = "local/sendgrid.jsonnet"
      }

      template {
        data = <<EOH
          DSN={{with secret "database-dev/creds/kratos"}}"postgres://{{.Data.username}}:{{.Data.password}}@localhost:5432/kratos?sslmode=disable"{{end}}
          {{- with secret "kv/data/dev/kratos/config" -}}
          SECRETS_DEFAULT={{.Data.data.SECRETS_DEFAULT}}
          SECRETS_COOKIE={{.Data.data.SECRETS_COOKIE}}
          SECRETS_CIPHER={{.Data.data.SECRETS_CIPHER}}
          COURIER_HTTP_REQUEST_CONFIG_AUTH_CONFIG_VALUE="Bearer {{.Data.data.SENDGRID_API_KEY}}"
          {{- end -}}
        EOH
        destination = "secrets/file.env"
        env = true
      }

      vault {}

      config {
        image   = "oryd/kratos:v1.0.0"
        args    = [
          "serve",
          "courier",
          "watch",
          "--config",
          "/local/kratos.yaml"
        ]
      }
    }
  }
}
