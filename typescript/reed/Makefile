
IMAGE_REPO := 823058932981.dkr.ecr.eu-west-1.amazonaws.com/reed
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

buildlocal: build
	@echo "Building local image"
	docker tag "$(IMAGE_REPO):$(IMAGE_HASH)" localhost:5005/reed
	@echo "Image built"

pushlocal: buildlocal
	@echo "Pushing image"
	docker push localhost:5005/reed
	@echo "Image pushed"
