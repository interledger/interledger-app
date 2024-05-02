#!/bin/bash

function install_nomad () {
	arch=$(lscpu | grep "Architecture" | awk '{print $NF}')
	if [[ $arch == x86_64* ]]; then
	  ARCH="amd64"
	elif  [[ $arch == aarch64 ]]; then
	  ARCH="arm64"
	fi
	echo -e '\e[38;5;198m'"CPU is $ARCH"
	echo -e '\e[38;5;198m'"Installing Nomad..."

	export DEBIAN_FRONTEND=noninteractive
	sudo apt-get update
	sudo apt-get install nomad

	# install and configure CNI plugin
	curl -L -o cni-plugins.tgz "https://github.com/containernetworking/plugins/releases/download/v1.0.0/cni-plugins-linux-$( [ $(uname -m) = aarch64 ] && echo arm64 || echo amd64)"-v1.0.0.tgz
	sudo mkdir -p /opt/cni/bin
  	sudo tar -C /opt/cni/bin -xzf cni-plugins.tgz

  	echo 1 | sudo tee /proc/sys/net/bridge/bridge-nf-call-arptables
    echo 1 | sudo tee /proc/sys/net/bridge/bridge-nf-call-ip6tables
    echo 1 | sudo tee /proc/sys/net/bridge/bridge-nf-call-iptables

  sudo mkdir --parents /etc/nomad.d
	sudo touch /etc/nomad.d/nomad.hcl
	sudo mkdir --parent /etc/nomad.d/certs/consul
	sudo cp /etc/consul.d/certs/consul-agent-ca.pem /etc/consul.d/certs/dc1-client-consul-0.pem /etc/consul.d/certs/dc1-client-consul-0-key.pem /etc/nomad.d/certs/consul
	sudo chown --recursive nomad:nomad /etc/nomad.d
	sudo chmod 640 /etc/nomad.d/nomad.hcl
	
	sudo mkdir -p /opt/nomad
	sudo chmod 750 /opt/nomad
	sudo mkdir -p /opt/nomad/data

	# make volume for postgres
	sudo mkdir -p /opt/nomad/data/volume/postgres
	sudo chown --recursive nomad:nomad /opt/nomad

cat << EOF | sudo tee /etc/nomad.d/server.hcl
data_dir  = "/opt/nomad/data"

bind_addr = "0.0.0.0" # the default

datacenter = "dc1"

advertise {
  # Defaults to the first private IP address.
  # http = "10.9.99.10"
  # rpc  = "10.9.99.10"
  # serf = "10.9.99.10:5648" # non-default ports may be specified
}

server {
  enabled          = true
  bootstrap_expect = 1
}

client {
  enabled       = true

  # https://www.nomadproject.io/docs/drivers/docker.html#volumes
  # https://github.com/hashicorp/nomad/issues/5562
  options = {
    "docker.volumes.enabled" = true
  }

  host_volume "postgres" {
    path      = "/opt/nomad/data/volume/postgres"
    read_only = false
  }
}

plugin "raw_exec" {
  config {
    enabled = true
  }
}

consul {
  address = "127.0.0.1:8501"
  ssl = true
  grpc_ca_file = "/etc/nomad.d/certs/consul/consul-agent-ca.pem"
  ca_file = "/etc/nomad.d/certs/consul/consul-agent-ca.pem"
  cert_file = "/etc/nomad.d/certs/consul/dc1-client-consul-0.pem"
  key_file = "/etc/nomad.d/certs/consul/dc1-client-consul-0-key.pem"
}

acl {
  enabled = true
}

vault {
  enabled = true
  address = "http://127.0.0.1:8200"

  default_identity {
    aud = ["vault.io"]
    ttl = "1h"
  }
}
EOF

    cat << EOF | sudo tee /etc/systemd/system/nomad.service
[Unit]
Description="HashiCorp Nomad - An orchestration tool"
Documentation=https://nomadproject.io/docs/
Wants=network-online.target
After=network-online.target

[Service]
ExecStart=/usr/bin/nomad agent -config=/etc/nomad.d/server.hcl
ExecReload=/bin/kill --signal HUP $MAINPID
KillMode=process
KillSignal=SIGTERM
Restart=on-failure
LimitNOFILE=65536

[Install]
WantedBy=multi-user.target
EOF

	sudo systemctl enable nomad
	sudo systemctl start nomad

    echo -e '\e[38;5;198m'"Done."
}

install_nomad
