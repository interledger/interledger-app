#!/bin/bash

# abort on unbound variable
set -o nounset

# don't hide errors within pipes
set -o pipefail

# exit on error
set -o errexit

REGISTRY="${REGISTRY:=localhost:5002}"
REPO="${REPO:=/home/vagrant/fynbos}"

function build_docker_images () {
	echo "Building temporal image..."
	docker build -t ${REGISTRY}/temporal -f ${REPO}/dev/nomad/temporal.Dockerfile ${REPO}/dev/nomad
	docker push ${REGISTRY}/temporal	

	echo "Building protea image..."
	docker build -t ${REGISTRY}/protea --target=dev -f ${REPO}/typescript/protea/Dockerfile ${REPO}/typescript/protea
	docker push ${REGISTRY}/protea

	echo "Building botanist image..."
	docker build -t ${REGISTRY}/botanist --target=dev -f ${REPO}/typescript/botanist/Dockerfile ${REPO}/typescript/botanist
	docker push ${REGISTRY}/botanist

	echo "Building backend image..."
	docker build -t ${REGISTRY}/backend --target=base -f ${REPO}/go/backend/Dockerfile ${REPO}/go
	docker push ${REGISTRY}/backend

	echo "Building mockbos image..."
	docker build -t ${REGISTRY}/mockbos --target=base -f ${REPO}/go/mockbos/Dockerfile ${REPO}/go
	docker push ${REGISTRY}/mockbos

	echo "Pulling rafiki backend"
	docker pull ghcr.io/interledger/rafiki-backend:v1.0.0-alpha.11
	docker tag ghcr.io/interledger/rafiki-backend:v1.0.0-alpha.11 localhost:5002/rafiki-backend
	docker push localhost:5002/rafiki-backend

	echo "Pulling rafiki auth"
	docker pull ghcr.io/interledger/rafiki-auth:v1.0.0-alpha.11
	docker tag ghcr.io/interledger/rafiki-auth:v1.0.0-alpha.11 localhost:5002/rafiki-auth
	docker push localhost:5002/rafiki-auth

	echo "Pulling rafiki frontend"
	docker pull ghcr.io/interledger/rafiki-frontend:v1.0.0-alpha.9
	docker tag ghcr.io/interledger/rafiki-frontend:v1.0.0-alpha.9 localhost:5002/rafiki-frontend
	docker push localhost:5002/rafiki-frontend
}

function deploy () {
	  # allow memory over subscription
	sudo apt install -y jq
	curl -s http://localhost:4646/v1/operator/scheduler/configuration | \
	  jq '.SchedulerConfig | .MemoryOversubscriptionEnabled=true' | \
	  curl -X PUT http://localhost:4646/v1/operator/scheduler/configuration -d @-
	
	cd $REPO/dev/nomad
	nomad job run -detach postgres.hcl
	nomad job run -detach redis.hcl
	nomad job run -detach temporal.hcl
	nomad job run -detach kratos.hcl
	nomad job run -detach backend.hcl
	nomad job run -detach mockbos.hcl
	nomad job run -detach protea.hcl
	nomad job run -detach botanist.hcl
	nomad job run -detach rafiki.hcl
	nomad job run -detach traefik.hcl
}

function install_node_modules() { 

    # TODO: Use Corepack and pin PNPM version as well across the projects
    sudo apt-get install npm
    npm install -g pnpm

    cd $REPO/typescript/protea
    pnpm install

    cd $REPO/botanist
    pnpm install
}

function main () {
    build_docker_images
    deploy
    install_node_modules
}

main "${@}"
