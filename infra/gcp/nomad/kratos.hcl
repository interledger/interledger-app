variable "dir" {
  type = string
}

job "kratos" {
  datacenters = [
    "dc1"
  ]
  type = "service"
  namespace = "dev"

  group "kratos" {
    count = 1

    update {
      max_parallel     = 1
      canary           = 1
      auto_revert      = true
      auto_promote     = true
      health_check     = "checks"
    }

    network {
      mode = "bridge"
      port "http" {
        to = 4433
      }
    }

    service {
      name = "dev-kratos-ingress"
      port = "http"

      tags = [
        "traefik.enable=true",
        "traefik.http.routers.dev-kratos.rule=(Host(`eu1.fynbos.dev`) && (PathPrefix(`/self-service`) || PathPrefix(`/sessions`)))"
      ]
    }

    service {
      name = "dev-kratos"
      port = 4433

      check { 
        name     = "alive"   
        type     = "http"
        port     = "http"
        path     = "/health/alive"
        interval = "5s"
        timeout  = "2s"
      }

      check {  
        name     = "ready"  
        type     = "http"
        port     = "http"
        path     = "/health/ready"
        interval = "5s"
        timeout  = "2s"
      }

      connect {
        sidecar_service {
          proxy {
            upstreams {
              destination_name = "dev-postgres"
              local_bind_port  = 5432
            }
            upstreams {
              destination_name = "dev-backend-http"
              local_bind_port  = 8080
            }
          }
        }
      }
    }

    service {
      name = "dev-kratos-admin"
      port = 4434
      connect {
        sidecar_service {}
      }
    }

    task "kratos-migrate" {
      driver = "docker"

      config {
        image   = "oryd/kratos:v1.1.0"
        args    = [
          "migrate",
          "sql",
          "-e",
          "--yes",
          "--config",
          "/etc/kratos/kratos.yaml"
        ]
        volumes = ["local/kratos.yaml:/etc/kratos/kratos.yaml"]
      }

      template {
        data = file(format("%s/kratos.yaml", var.dir))
        destination = "local/kratos.yaml"
      }

      template {
        data = <<EOH
          {{with secret "database-dev/static-creds/kratos"}}
          DSN="postgres://{{.Data.username}}:{{.Data.password}}@localhost:5432/kratos?sslmode=disable"
          {{end}}
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
        data = file(format("%s/kratos.yaml", var.dir))
        destination = "local/kratos.yaml"
      }

      template {
        data = file(format("%s/kratos-identity.json", var.dir))
        destination = "local/identity.schema.json"
      }

      template {
        data = file(format("%s/kratos-sendgrid.jsonnet", var.dir))
        destination = "local/sendgrid.jsonnet"
      }

      template {
        data = <<EOH
          {{with secret "database-dev/static-creds/kratos"}}
          DSN="postgres://{{.Data.username}}:{{.Data.password}}@localhost:5432/kratos?sslmode=disable&max_conns=4&max_idle_conns=4&connect_timeout=5"
          {{end}}
          
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
        image   = "oryd/kratos:v1.1.0"
        args    = [
          "serve",
          "all",
          "--watch-courier",
          "--config",
          "/etc/kratos/kratos.yaml"
        ]
        volumes = [
          "local/kratos.yaml:/etc/kratos/kratos.yaml",
          "local/sendgrid.jsonnet:/etc/kratos/sendgrid.jsonnet",
          "local/identity.schema.json:/etc/kratos/identity.schema.json"
        ]
      }

      resources {
        cpu = 500
        memory = 2048
      }
    }

    # task "kratos-mail" {
    #   driver = "docker"

    #   template {
    #     data = file(format("%s/kratos.yaml", var.dir))
    #     destination = "local/kratos.yaml"
    #   }

    #   template {
    #     data = file(format("%s/kratos-identity.json", var.dir))
    #     destination = "local/identity.schema.json"
    #   }

    #   template {
    #     data = file(format("%s/kratos-sendgrid.jsonnet", var.dir))
    #     destination = "local/sendgrid.jsonnet"
    #   }

    #   template {
    #     data = <<EOH
    #       {{with secret "database-dev/static-creds/kratos"}}
    #       DSN="postgres://{{.Data.username}}:{{.Data.password}}@localhost:5432/kratos?sslmode=disable"
    #       {{end}}
          
    #       {{- with secret "kv/data/dev/kratos/config" -}}
    #       SECRETS_DEFAULT={{.Data.data.SECRETS_DEFAULT}}
    #       SECRETS_COOKIE={{.Data.data.SECRETS_COOKIE}}
    #       SECRETS_CIPHER={{.Data.data.SECRETS_CIPHER}}
    #       COURIER_HTTP_REQUEST_CONFIG_AUTH_CONFIG_VALUE="Bearer {{.Data.data.SENDGRID_API_KEY}}"
    #       {{- end -}}
    #     EOH
    #     destination = "secrets/file.env"
    #     env = true
    #   }

    #   vault {}

    #   config {
    #     image   = "oryd/kratos:v1.0.0"
    #     args    = [
    #       "serve",
    #       "courier",
    #       "watch",
    #       "--config",
    #       "/local/kratos.yaml"
    #     ]
    #   }
    # }
  }
}
