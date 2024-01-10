job "kratos" {
  datacenters = [
    "dc1"
  ]
  type = "service"

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

    task "kratos" {
      driver = "docker"

      template {
        data = file("kratos.yml")
        destination = "local/kratos.yml"
      }

      template {
        data = file("identity.json")
        destination = "local/identity.schema.json"
      }

      config {
        image   = "oryd/kratos:v1.0.0"
        args    = [
          "serve",
          "all",
          "--dev",
          "--watch-courier",
          "--config",
          "/local/kratos.yml"
        ]
      }

      resources {
        cpu    = 100
        memory = 256
      }
    }
  }
}
