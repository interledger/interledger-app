#!/bin/bash


# abort on unbound variable
set -o nounset

# don't hide errors within pipes
set -o pipefail

# Check if the environment file exists, error out if it does not
if [ ! -f environment ]; then
  echo "Environment file not found. Exiting..."
  exit 1
fi
source environment

# GCP KMS credentials used as vault backend
GCP_KMS_PROJECT=${GCP_KMS_PROJECT:=wallet-451208}
GCP_KMS_REGION=${GCP_KMS_REGION:=global}
GCP_KMS_KEY_RING=${GCP_KMS_KEY_RING:=vault-keys}
GCP_KMS_CRYPTO_KEY=${GCP_KMS_CRYPTO_KEY:=vault-unseal-key}

# If HOST_IP is not set, the script will fail
if [ -z "${HOST_IP}" ]; then
  echo "HOST_IP is not set. Exiting..."
  exit 1
fi

# Even though the original authors tried to use sudo everywhere there are still some parts
# that will fail if we are not root. Perhaps they ran a different distro.
# If not root then fail
if [ "$(id -u)" -ne 0 ]; then
  echo "This script must be run as root" 1>&2
  exit 1
fi


# HOST details
HOST_IP="${HOST_IP:=}"

function add_remote_repositories() {
	echo "Updating repositories..."
	sudo apt-get update
	sudo apt-get --assume-yes install curl gnupg lsb-release ca-certificates < /dev/null > /dev/null

	echo "Adding hashicorp and docker repositories..."
	local arch=$(lscpu | grep "Architecture" | awk '{print $NF}')
	local ARCH
	if [[ $arch == x86_64* ]]; then
	  ARCH="amd64"
	elif  [[ $arch == aarch64 ]]; then
	  ARCH="arm64"
	fi
	echo -e '\e[38;5;198m'"CPU is ${ARCH}"

	export DEBIAN_FRONTEND=noninteractive

	# add hashicorp gpg key
	curl --fail --silent --show-error --location https://apt.releases.hashicorp.com/gpg | \
      gpg --dearmor | \
      sudo dd of=/usr/share/keyrings/hashicorp-archive-keyring.gpg

    # add hashicorp linux repo
    echo "deb [arch=amd64 signed-by=/usr/share/keyrings/hashicorp-archive-keyring.gpg] https://apt.releases.hashicorp.com $(lsb_release -cs) main" | \
	sudo tee -a /etc/apt/sources.list.d/hashicorp.list

	# add docker key
	sudo install -m 0755 -d /etc/apt/keyrings
	sudo curl -fsSL https://download.docker.com/linux/debian/gpg -o /etc/apt/keyrings/docker.asc
	sudo chmod a+r /etc/apt/keyrings/docker.asc

	# add docker linux repo
	echo \
	  "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.asc] https://download.docker.com/linux/debian \
	  $(. /etc/os-release && echo "$VERSION_CODENAME") stable" | \
	  sudo tee /etc/apt/sources.list.d/docker.list > /dev/null

	# update repositories  
	sudo apt-get update	
}

function install_docker () {
	echo "Installing docker..."
	export DEBIAN_FRONTEND=noninteractive
	sudo apt-get install --assume-yes docker-ce docker-ce-cli containerd.io docker-buildx-plugin docker-compose-plugin

	sudo systemctl enable docker.service
	sudo systemctl enable containerd.service
}

function install_vault () {
	echo "Installing Vault..."
	export DEBIAN_FRONTEND=noninteractive
	sudo apt-get install vault

	# allow vault to use mlock
	sudo setcap cap_ipc_lock=+ep /usr/bin/vault

	echo "Creating Vault config files..."
	sudo mkdir --parents /etc/vault.d
	sudo touch /etc/vault.d/vault.hcl
	sudo chown --recursive vault:vault /etc/vault.d
	sudo chmod 640 /etc/vault.d/vault.hcl

	cat << VAULTCONFIG | sudo tee /etc/vault.d/vault.hcl
storage "file" {
  path    = "/data/live/vault/data"
}

listener "tcp" {
  address     = "127.0.0.1:8200"
  tls_disable = "true"
}

seal "gcpckms" {
  project     = "${GCP_KMS_PROJECT}"
  region      = "${GCP_KMS_REGION}"
  key_ring    = "${GCP_KMS_KEY_RING}"
  crypto_key  = "${GCP_KMS_CRYPTO_KEY}"
}

api_addr = "http://127.0.0.1:8200"
ui = true
VAULTCONFIG

	echo "Creating Vault data directory..."
	sudo mkdir -p /data/live/vault
	sudo mkdir -p /data/live/vault/data
	sudo chown --recursive vault:vault /data/live/vault

	echo "Creating Vault systemd config..."
	cat << 'VAULTSERVICE' | sudo tee /etc/systemd/system/vault.service
[Unit]
Description="HashiCorp Vault - A tool for managing secrets"
Documentation=https://www.vaultproject.io/
Requires=network-online.target
After=network-online.target
ConditionFileNotEmpty=/etc/vault.d/vault.hcl

[Service]
User=vault
Group=vault
ExecStart=/usr/bin/vault server -config=/etc/vault.d/vault.hcl
ExecReload=/bin/kill --signal HUP
KillMode=process
KillSignal=SIGTERM
Restart=on-failure
LimitNOFILE=65536

[Install]
WantedBy=multi-user.target
VAULTSERVICE

	sudo systemctl enable vault
}

function install_consul () {
	echo "Installing Consul..."
	export DEBIAN_FRONTEND=noninteractive
	sudo apt-get install consul

	echo "Creating Consul config files..."
	sudo mkdir --parents /etc/consul.d
	sudo mkdir --parents /etc/consul.d/certs
	sudo touch /etc/consul.d/server.hcl
	(cd /etc/consul.d/certs && consul tls ca create)
	(cd /etc/consul.d/certs && consul tls cert create -server)
	(cd /etc/consul.d/certs && consul tls cert create -client)
	sudo chown --recursive consul:consul /etc/consul.d
	sudo chmod 640 /etc/consul.d/server.hcl

	local consul_encryption_key=$(consul keygen)
	cat << CONSULCONFIG | sudo tee /etc/consul.d/server.hcl
datacenter = "dc1"
data_dir = "/opt/consul"
encrypt = "${consul_encryption_key}"
log_level = "INFO"
server = true
bootstrap_expect = 1

client_addr = "127.0.0.1"
bind_addr = "127.0.0.1"
advertise_addr = "127.0.0.1"

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

acl {
  enabled = false
  default_policy = "allow"
#   enable_token_persistence = true
}

connect {
  enabled = true
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
CONSULCONFIG

	echo "Creating Consul data directory..."
	sudo mkdir /opt/consul
	sudo chown --recursive consul:consul /opt/consul
	sudo chmod 750 /opt/consul

	echo "Creating Consul systemd config..."
	cat << 'CONSULSERVICE' | sudo tee /etc/systemd/system/consul.service
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
ExecReload=/bin/kill --signal HUP
KillMode=process
KillSignal=SIGTERM
Restart=on-failure
LimitNOFILE=65536

[Install]
WantedBy=multi-user.target
CONSULSERVICE

	sudo systemctl enable consul
}

function install_nomad () {
	export DEBIAN_FRONTEND=noninteractive
	sudo apt-get install nomad

	echo "Install and configure Nomad CNI..."
	curl -L -o cni-plugins.tgz "https://github.com/containernetworking/plugins/releases/download/v1.0.0/cni-plugins-linux-$( [ $(uname -m) = aarch64 ] && echo arm64 || echo amd64)"-v1.0.0.tgz
	sudo mkdir -p /opt/cni/bin
  	sudo tar -C /opt/cni/bin -xzf cni-plugins.tgz

  	echo 1 | sudo tee /proc/sys/net/bridge/bridge-nf-call-arptables
    echo 1 | sudo tee /proc/sys/net/bridge/bridge-nf-call-ip6tables
    echo 1 | sudo tee /proc/sys/net/bridge/bridge-nf-call-iptables

    echo "Creating Consul config files..."
    sudo mkdir --parents /etc/nomad.d
	sudo touch /etc/nomad.d/nomad.hcl
	sudo mkdir --parent /etc/nomad.d/certs/consul
	sudo cp /etc/consul.d/certs/consul-agent-ca.pem /etc/consul.d/certs/dc1-client-consul-0.pem /etc/consul.d/certs/dc1-client-consul-0-key.pem /etc/nomad.d/certs/consul
	sudo chown --recursive nomad:nomad /etc/nomad.d
	sudo chmod 640 /etc/nomad.d/nomad.hcl

	cat << NOMADCONFIG | sudo tee /etc/nomad.d/server.hcl
data_dir  = "/data/live/nomad/data"

bind_addr = "${HOST_IP}"

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

  host_volume "db-prod" {
    path      = "/data/live/nomad/data/volume/db-prod"
    read_only = false
  }

  host_volume "db-dev" {
    path      = "/data/live/nomad/data/volume/db-dev"
    read_only = false
  }
}

plugin "raw_exec" {
  config {
    enabled = true
  }
}

# Fynbos use a credential helper to pull images from ECR
#
# plugin "docker" {
#   config {
#     allow_privileged = true
#     auth {
#       config = "/etc/nomad.d/.docker/config.json"
#     }
#     volumes {
#      enabled = true
#     }
#   }
# }

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
NOMADCONFIG
	
	echo "Creating Nomad data directories..."
	# make volume for nomad data. Assumes there is a zfs file system mounted at /data
	sudo mkdir -p /data/live/nomad
	sudo chmod 750 /data/live/nomad
	sudo mkdir -p /data/live/nomad/data

	# make volume for dev and prod postgres instances
	sudo mkdir -p /data/live/nomad/data/volume/db-prod
	sudo mkdir -p /data/live/nomad/data/volume/db-dev
	sudo chown --recursive nomad:nomad /data/live/nomad

cat << NOMADSERVICE | sudo tee /etc/systemd/system/nomad.service
[Unit]
Description="HashiCorp Nomad - An orchestration tool"
Documentation=https://nomadproject.io/docs/
Wants=network-online.target
After=network-online.target

[Service]
ExecStart=/usr/bin/nomad agent -config=/etc/nomad.d/server.hcl
ExecReload=/bin/kill --signal HUP
KillMode=process
KillSignal=SIGTERM
Restart=on-failure
LimitNOFILE=65536

[Install]
WantedBy=multi-user.target
NOMADSERVICE

	sudo systemctl enable nomad
}

function main () {
	add_remote_repositories
	install_docker
	install_vault
	install_consul
	install_nomad

	sudo systemctl start vault
	sudo systemctl start consul
	sudo systemctl start nomad

}

main "${@}"
