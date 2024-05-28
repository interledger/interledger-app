job "consul-server" {
  datacenters = [
    "dc1"
  ]
  type        = "system"
  priority    = 100

  group "consul" {

    restart {
      mode     = "delay"
      interval = "1m"
    }

    task "consul" {
      driver = "docker"

      config {
        network_mode = "host"
        image        = "hashicorp/consul:1.16"
        args         = [
          "agent",
          "-config-file",
          "/local/consul.hcl"
        ]

        mount {
          type = "volume"
          source = "opt-consul"
          target = "/opt/consul"
        }
      }

      template {
        data            = <<EOF
          performance {
            raft_multiplier = 1
          }

          # Set our node_name to a fixed value to make TLS much easier. If
          # we have multiple servers per DC, we'll probably need to make this
          # more dynamic.
          node_name          = "consul-server"
          server             = true
          data_dir           = "/opt/consul"
          client_addr        = "127.0.0.1"
          bind_addr          = "127.0.0.1"
          bootstrap_expect   = 1
          datacenter         = "dc1"

          ports {
            grpc = 8502
          }

          connect {
            enabled                 = true
          }

          ui = true
        EOF
        left_delimiter  = "{{{"
        right_delimiter = "}}}"
        destination     = "local/consul.hcl"
      }

      resources {
        cpu    = 100
        memory = 128
      }
    }
  }
}
