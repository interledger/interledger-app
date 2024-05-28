job "traefik" {
  datacenters = ["dc1"]
  type        = "service"
  namespace   = "default"

  group "traefik" {
    count = 1

    network {
      port  "http"{
        static = 80
      }
      port  "https"{
        static = 443
      }
      port  "admin"{
        static = 8080
      }
    }

    service {
      name = "traefik"
      check {
        name     = "alive"
        type     = "tcp"
        port     = "http"
        interval = "10s"
        timeout  = "2s"
      }
    }

    task "traefik" {
      driver = "docker"
      config {
        image = "traefik:2.10"
        ports = ["admin", "http"]
        network_mode = "host"
        volumes = [
          "local/traefik.toml:/etc/traefik/traefik.toml",
          "local/config.toml:/etc/traefik/config.toml",
          "secrets/tls.cert:/etc/traefik/tls.cert",
          "secrets/tls.key:/etc/traefik/tls.key"
        ]
      }

      vault {}

      template {
        data = <<EOF
        {{with secret "kv/data/default/traefik/config"}}
        {{.Data.data.cert}}
        {{end}}
        EOF
        destination = "secrets/tls.cert"
      }

      template {
        data = <<EOF
        {{with secret "kv/data/default/traefik/config"}}
        {{.Data.data.key}}
        {{end}}
        EOF
        destination = "secrets/tls.key"
      }

      template {
        data = <<EOF
[entryPoints]
    [entryPoints.web]
      address = ":80"
        [entryPoints.web.http]
          [entryPoints.web.http.redirections]
            [entryPoints.web.http.redirections.entryPoint]
              to = "websecure"
              scheme = "https"
    [entryPoints.websecure]
      address = ":443"
        [entryPoints.websecure.http.tls]
    [entryPoints.traefik]
      address = ":8080"

[log]
  level = "DEBUG"

[api]
    dashboard = true
    insecure  = true

# Enable Consul Catalog configuration backend.
[providers.consulCatalog]
    prefix           = "traefik"
    exposedByDefault = false

    [providers.consulCatalog.endpoint]
      address = "127.0.0.1:8500"
      scheme  = "http"

[providers.file]
    filename = "/etc/traefik/config.toml"
    watch = true
EOF

        destination = "local/traefik.toml"
      }

      template {
        data = <<EOF
[tls]
  [[tls.certificates]]
    certFile = "/etc/traefik/tls.cert"
    keyFile = "/etc/traefik/tls.key"
EOF
        
        destination = "local/config.toml"
      }

    }
  }
}
