data_dir = "/var/lib/nomad"

bind_addr = "172.17.0.1" # Use docker network interface

server {
  enabled          = true
  bootstrap_expect = 1
}

client {
  enabled = true
  host_volume "opt-consul" {
    path      = "/opt/consul"
    read_only = false
  }

  host_volume "postgres" {
    path      = "/opt/postgres/data"
    read_only = false
  }
}

plugin "docker" {
  enabled = true
  config {
    volumes {
      enabled = true
    }
    gc {
      image = false
    }
  }
}