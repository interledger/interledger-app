
IMAGE_REPO := 823058932981.dkr.ecr.eu-west-1.amazonaws.com/botanist
IMAGE_HASH := latest

test:
	@echo "No tests to run"

format:
	@echo "No formatting"

lint:
	@echo "No Linting"

build:
	@echo "Building image"
	DOCKER_BUILDKIT=1 docker build -t "$(IMAGE_REPO):$(IMAGE_HASH)" -f ./Dockerfile .
	@echo "Image built"

push:
	@echo "Pushing image"
	docker push "$(IMAGE_REPO):$(IMAGE_HASH)"
	@echo "Image pushed"

devdeploy:
	@echo "Deploying to local environment..."
	(cd ../../dev/vagrant && vagrant ssh -c 'nomad job run /home/vagrant/fynbos/dev/nomad/botanist.hcl')
