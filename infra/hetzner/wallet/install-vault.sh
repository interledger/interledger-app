#!/bin/bash

# https://developer.hashicorp.com/vault/tutorials/getting-started/getting-started-install

function install_vault () {
	arch=$(lscpu | grep "Architecture" | awk '{print $NF}')
	if [[ $arch == x86_64* ]]; then
	  ARCH="amd64"
	elif  [[ $arch == aarch64 ]]; then
	  ARCH="arm64"
	fi
	echo -e '\e[38;5;198m'"CPU is $ARCH"
	echo -e '\e[38;5;198m'"Installing Vault..."

	export DEBIAN_FRONTEND=noninteractive
	sudo apt-get update
	sudo apt-get install vault

	# set up data and config directories
	sudo mkdir --parents /etc/vault.d
	sudo touch /etc/vault.d/vault.hcl
	sudo chown --recursive vault:vault /etc/vault.d
	sudo chmod 640 /etc/vault.d/vault.hcl
	
	  # make volume for vault data. Assumes there is a zfs file system mounted at /data
	sudo mkdir -p /data/live/vault
	sudo mkdir -p /data/live/vault/data
	sudo chown --recursive vault:vault /data/live/vault
	
	# allow vault to use mlock and create vault user
	sudo setcap cap_ipc_lock=+ep /usr/bin/vault

	cat << EOF | sudo tee /etc/vault.d/vault.hcl
storage "file" {
  path    = "/data/live/vault/data"
}

listener "tcp" {
  address     = "127.0.0.1:8200"
  tls_disable = "true"
}

seal "awskms" {
  region     = "eu-west-1"
  access_key = ""
  secret_key = ""
  kms_key_id = ""
}

api_addr = "http://127.0.0.1:8200"
ui = true
EOF

# TODO: unseal using aws kms key

	cat << EOF | sudo tee /etc/systemd/system/vault.service
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
ExecReload=/bin/kill --signal HUP $MAINPID
KillMode=process
KillSignal=SIGTERM
Restart=on-failure
LimitNOFILE=65536

[Install]
WantedBy=multi-user.target
EOF

	sudo systemctl enable vault
	sudo systemctl start vault

	echo -e '\e[38;5;198m'"Done."
}

install_vault
