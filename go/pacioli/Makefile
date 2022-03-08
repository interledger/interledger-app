
IMAGE_REPO := 823058932981.dkr.ecr.eu-west-1.amazonaws.com/pacioli
IMAGE_HASH := latest

pull:
	@echo "Pulling required images"
	docker pull 823058932981.dkr.ecr.eu-west-1.amazonaws.com/cockroach:latest-v21.1
	docker pull 823058932981.dkr.ecr.eu-west-1.amazonaws.com/tigerbeetle
	@echo "Pull complete"

test:
	@echo "Running tests"
	go test ./...
	@echo "Tests complete"

format:
	@echo "Formatting"
	go fmt ./...
	@echo "Formate complete"

lint:
	@echo "Linting"
	golangci-lint run ./...
	@echo "Lint complete"

build:
	@echo "Building image"
	DOCKER_BUILDKIT=1 docker build -t "$(IMAGE_REPO):$(IMAGE_HASH)" -f ./Dockerfile ./../
	@echo "Image built"

push:
	@echo "Pushing image"
	docker push "$(IMAGE_REPO):$(IMAGE_HASH)"
	@echo "Image pushed"