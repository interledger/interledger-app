#!/bin/bash

# abort on unbound variable
set -o nounset

# don't hide errors within pipes
set -o pipefail

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

	wget -O- https://apt.releases.hashicorp.com/gpg | \
  		sudo gpg --dearmor -o /usr/share/keyrings/hashicorp-archive-keyring.gpg

	echo "deb [signed-by=/usr/share/keyrings/hashicorp-archive-keyring.gpg] https://apt.releases.hashicorp.com $(lsb_release -cs) main" \
	    | sudo tee /etc/apt/sources.list.d/hashicorp.list

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

	sudo usermod -aG docker vagrant
	sudo mkdir -p /etc/docker

	cat << CONFIG | sudo tee /etc/docker/daemon.json
{
  "metrics-addr": "0.0.0.0:9323",
  "experimental": true,
  "storage-driver": "overlay2",
  "insecure-registries": ["10.9.99.10:5001", "10.9.99.10:5002", "localhost:5001", "localhost:5002"]
}
CONFIG
	sudo service docker restart

  # username=admin, password=password
  sudo mkdir -p /etc/docker/auth
  cat << 'HTPASSWD' | sudo tee /etc/docker/auth/htpasswd
admin:$2y$05$sfNkbq1WiINvqsE1wIeeEudN9MwnPay9YcprK69A4UCRYI37XyaYC
HTPASSWD

cat << AUTHCONFIG | sudo tee /etc/docker/auth.json
{
  "username": "admin",
  "password": "password",
  "email": "admin@localhost"
}
AUTHCONFIG

echo "Starting docker registry..."
# https://docs.docker.com/registry/deploying/#customize-the-published-port
docker run -d --restart=always \
  --name registry \
  -v /etc/docker/auth/htpasswd:/auth/htpasswd \
  -e "REGISTRY_AUTH=htpasswd" \
  -e "REGISTRY_AUTH_HTPASSWD_REALM=Registry Realm" \
  -e REGISTRY_AUTH_HTPASSWD_PATH=/auth/htpasswd \
  -e REGISTRY_HTTP_ADDR=0.0.0.0:5002 \
  --memory 256M -p 5002:5002 registry:2

	# authenticate vagrant user with local registry
	mkdir -p /home/vagrant/.docker
	cat << DOCKERCONFIG | sudo tee /home/vagrant/.docker/config.json
{
	"auths": {
		"localhost:5002": {
			"auth": "YWRtaW46cGFzc3dvcmQ="
		}
	}
}
DOCKERCONFIG
	sudo chown -R vagrant:vagrant /home/vagrant/.docker
	sudo chmod 755 /home/vagrant/.docker/config.json

	sudo systemctl enable docker.service
	sudo systemctl enable containerd.service
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

client_addr = "0.0.0.0"
bind_addr = "0.0.0.0"
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
	sudo systemctl start consul
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
data_dir  = "/opt/nomad/data"

bind_addr = "0.0.0.0"

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

  host_volume "go" {
    path      = "/home/vagrant/fynbos/go"
    read_only = false
  }

  host_volume "protea" {
    path      = "/home/vagrant/fynbos/typescript/protea"
    read_only = false
  }

  host_volume "botanist" {
    path      = "/home/vagrant/fynbos/typescript/botanist"
    read_only = false
  }
}

plugin "docker" {
  config {
    auth {
      config = "/home/vagrant/.docker/config.json"
    }

    volumes {
    	enabled = true
    }

    allow_privileged = true
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
  enabled = false
}
NOMADCONFIG
    
	echo "Creating Nomad data directories..."
	sudo mkdir -p /opt/nomad
	sudo chmod 750 /opt/nomad
	sudo mkdir -p /opt/nomad/data/volume/postgres
	sudo chown --recursive nomad:nomad /opt/nomad

cat << NOMADSERVICE | sudo tee /etc/systemd/system/nomad.service
[Unit]
Description="HashiCorp Nomad - An orchestration tool"
Documentation=https://nomadproject.io/docs/
Wants=network-online.target consul.service
After=network-online.target consul.service
Requires=consul.service

[Service]
ExecStartPre=/usr/bin/bash -c "until curl -s -k https://127.0.0.1:8501/v1/status/leader | grep -q .; do echo 'Waiting for Consul HTTPS API...'; sleep 1; done"
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
	sudo systemctl start nomad
}

function install_virtualbox_guest_additions () {
	sudo apt-get -y install linux-headers-$(uname -r) build-essential dkms

	wget https://download.virtualbox.org/virtualbox/7.0.14/VBoxGuestAdditions_7.0.14.iso
	sudo mkdir /media/VBoxGuestAdditions
	sudo mount -o loop,ro VBoxGuestAdditions_7.0.14.iso /media/VBoxGuestAdditions
	sudo sh /media/VBoxGuestAdditions/VBoxLinuxAdditions.run
	rm VBoxGuestAdditions_7.0.14.iso
	sudo umount /media/VBoxGuestAdditions
	sudo rmdir /media/VBoxGuestAdditions
}

function main () {
	add_remote_repositories
	install_virtualbox_guest_additions
	install_docker
	install_consul
	install_nomad
}

main "${@}"
