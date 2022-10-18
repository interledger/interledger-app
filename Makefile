
current_dir := $(shell dirname $(realpath $(firstword $(MAKEFILE_LIST))))

kindup:
	@echo "Bringing Cluster Online"
	./dev/kind/run.sh
	@echo "Cluster online. Deploying..."
	./dev/kind/deploy.sh

kinddeploy:
	@echo "Deploying..."
	./dev/kind/deploy.sh

kinddown:
	@echo "Deleting Cluster"
	./dev/kind/nuke.sh
	@echo "Cluster deleted"

buildgo:
	DOCKER_BUILDKIT=1 docker build $(current_dir)/go -f $(current_dir)/go/$(target)/Dockerfile -t localhost:5005/$(target):latest
	docker push localhost:5005/$(target):latest

tiltup:
	@echo "Running tilt"
	tilt up -f ./dev/tilt/Tiltfile
