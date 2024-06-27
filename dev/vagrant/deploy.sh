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
	nomad job run -detach protea.hcl
	nomad job run -detach botanist.hcl
	nomad job run -detach traefik.hcl
}

function main () {
	build_docker_images
	deploy
}

main "${@}"
