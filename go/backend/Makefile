
IMAGE_REPO := 823058932981.dkr.ecr.eu-west-1.amazonaws.com/backend
IMAGE_HASH := latest

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
	go vet ./...
	@echo "Lint complete"

build:
	@echo "Building image"
	DOCKER_BUILDKIT=1 docker build -t "$(IMAGE_REPO):$(IMAGE_HASH)" -f ./Dockerfile ./../
	@echo "Image built"

push:
	@echo "Pushing image"
	docker push "$(IMAGE_REPO):$(IMAGE_HASH)"
	@echo "Image pushed"