
current_dir := $(shell dirname $(realpath $(firstword $(MAKEFILE_LIST))))

kindup:
	@echo "Bringing Cluster Online"
	./dev/kind/run.sh
	@echo "Cluster online. Deploying..."
	./dev/kind/deploy.sh
	@echo "Adding localCA to trust store"
	./dev/kind/installCA.sh

kinddeploy:
	@echo "Deploying..."
	./dev/kind/deploy.sh

kinddown:
	@echo "Deleting Cluster"
	./dev/kind/nuke.sh
	@echo "Cluster deleted"

kindupdateca:
	@echo "Adding localCA to trust store"
	./dev/kind/installCA.sh

buildgo:
	DOCKER_BUILDKIT=1 docker build $(current_dir)/go -f $(current_dir)/go/$(target)/Dockerfile -t localhost:5005/$(target):latest
	docker push localhost:5005/$(target):latest

tiltup:
	@echo "Running tilt"
	tilt up -f ./dev/tilt/Tiltfile

devup:
	@echo "Installing local CA..."
	(cd ./dev/vagrant && ./installCA.sh)
	@echo "Bringing local environment online..."
	(cd ./dev/vagrant && vagrant up)
	@echo "Deploying to local environment..."
	(cd ./dev/vagrant && vagrant reload && vagrant ssh < deploy.sh)
	@echo "Done."

devdeploy:
	@echo "Deploying to local environment..."
	(cd ./dev/vagrant && vagrant reload && vagrant ssh < deploy.sh)
	@echo "Done."

devdown:
	@echo "Deleting local environment..."
	(cd ./dev/vagrant && vagrant destroy)
	@echo "Done."

localwallet:
	FYNBOS_ENV=local go run go/backend/cli/dev/main.go make wallet -k -l
