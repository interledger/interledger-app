#!/bin/bash

function install_docker () {
	arch=$(lscpu | grep "Architecture" | awk '{print $NF}')
	if [[ $arch == x86_64* ]]; then
	  ARCH="amd64"
	elif  [[ $arch == aarch64 ]]; then
	  ARCH="arm64"
	fi
	echo -e '\e[38;5;198m'"CPU is $ARCH"
	echo -e '\e[38;5;198m'"Installing Docker..."

	export DEBIAN_FRONTEND=noninteractive
	sudo apt-get update
	sudo apt-get install ca-certificates curl
	sudo install -m 0755 -d /etc/apt/keyrings
	sudo curl -fsSL https://download.docker.com/linux/debian/gpg -o /etc/apt/keyrings/docker.asc
	sudo chmod a+r /etc/apt/keyrings/docker.asc

	# Add the repository to Apt sources:
	echo \
	  "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.asc] https://download.docker.com/linux/debian \
	  $(. /etc/os-release && echo "$VERSION_CODENAME") stable" | \
	  sudo tee /etc/apt/sources.list.d/docker.list > /dev/null
	sudo apt-get update

	sudo apt-get install --assume-yes docker-ce docker-ce-cli containerd.io docker-buildx-plugin docker-compose-plugin

	sudo groupadd docker
	sudo usermod -aG docker $USER
	sudo newgrp docker

	sudo systemctl enable docker.service
	sudo systemctl enable containerd.service

	echo -e '\e[38;5;198m'"Installing Done."
}

install_docker
