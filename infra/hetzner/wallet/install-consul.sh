#!/bin/bash

# https://developer.hashicorp.com/consul/tutorials/production-vms/deployment-guide

function install_consul () {
	arch=$(lscpu | grep "Architecture" | awk '{print $NF}')
	if [[ $arch == x86_64* ]]; then
	  ARCH="amd64"
	elif  [[ $arch == aarch64 ]]; then
	  ARCH="arm64"
	fi
	echo -e '\e[38;5;198m'"CPU is $ARCH"
	echo -e '\e[38;5;198m'"Installing Consul..."

	export DEBIAN_FRONTEND=noninteractive
	sudo apt-get update
	sudo apt-get install consul

	# encryption key for gossip
	sudo mkdir --parents /etc/consul.d
	consul keygen | sudo tee /etc/consul.d/encryption_key.txt

	# setup TLS certs for rpc
	consul tls ca create
	consul tls cert create -server
	consul tls cert create -client
	sudo mkdir /etc/consul.d/certs && sudo mv ./*.pem /etc/consul.d/certs

	sudo mkdir /opt/consul
	sudo chown --recursive consul:consul /opt/consul
	sudo chmod 750 /opt/consul
	cat <<EOF | sudo tee /etc/consul.d/server.hcl
datacenter = "dc1"
data_dir = "/opt/consul"
encrypt = "$(cat /etc/consul.d/encryption_key.txt)"
log_level = "INFO"
server = true
bootstrap_expect = 1

client_addr = "127.0.0.1"
bind_addr = "10.0.0.10"
advertise_addr = "10.0.0.10"

tls {
   defaults {
      ca_file = "/etc/consul.d/certs/consul-agent-ca.pem"
      cert_file = "/etc/consul.d/certs/dc1-server-consul-0.pem"
      key_file = "/etc/consul.d/certs/dc1-server-consul-0-key.pem"

      # verify_incoming = true
      verify_outgoing = true
   }
   internal_rpc {
      verify_server_hostname = true
   }
}

auto_encrypt {
  allow_tls = true
}

# acl {
#   enabled = true
#   default_policy = "deny"
#   enable_token_persistence = true
# }

connect {
  enabled = true
}

addresses {

}

ports {
  grpc_tls  = 8502
  dns   = 8600
  http  = 8500
  https = 8501
}

ui_config {
  enabled = true
}

EOF
	sudo chown --recursive consul:consul /etc/consul.d
	sudo chmod 640 /etc/consul.d/server.hcl

	cat << EOF | sudo tee /etc/systemd/system/consul.service
[Unit]
Description="HashiCorp Consul - A service mesh solution"
Documentation=https://www.consul.io/
Requires=network-online.target
After=network-online.target
ConditionFileNotEmpty=/etc/consul.d/server.hcl

[Service]
EnvironmentFile=-/etc/consul.d/consul.env
User=consul
Group=consul
ExecStart=/usr/bin/consul agent -config-dir=/etc/consul.d/ -config-file=/etc/consul.d/server.hcl
ExecReload=/bin/kill --signal HUP $MAINPID
KillMode=process
KillSignal=SIGTERM
Restart=on-failure
LimitNOFILE=65536

[Install]
WantedBy=multi-user.target
EOF

	sudo systemctl enable consul
	sudo systemctl start consul

	echo -e '\e[38;5;198m'"Done."
}

install_consul
