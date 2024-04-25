#!/bin/bash

function add_hashicorp_repo() {
	arch=$(lscpu | grep "Architecture" | awk '{print $NF}')
	if [[ $arch == x86_64* ]]; then
	  ARCH="amd64"
	elif  [[ $arch == aarch64 ]]; then
	  ARCH="arm64"
	fi
	echo -e '\e[38;5;198m'"CPU is $ARCH"

	export DEBIAN_FRONTEND=noninteractive

	sudo apt-get update
	sudo apt-get --assume-yes install curl gnupg lsb-release < /dev/null > /dev/null

	# add hashicorp gpg key
	curl --fail --silent --show-error --location https://apt.releases.hashicorp.com/gpg | \
      gpg --dearmor | \
      sudo dd of=/usr/share/keyrings/hashicorp-archive-keyring.gpg

    # add hashicorp linux repo
    echo "deb [arch=amd64 signed-by=/usr/share/keyrings/hashicorp-archive-keyring.gpg] https://apt.releases.hashicorp.com $(lsb_release -cs) main" | \
	sudo tee -a /etc/apt/sources.list.d/hashicorp.list

	sudo apt-get update
}

add_hashicorp_repo
