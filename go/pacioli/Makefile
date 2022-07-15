
IMAGE_REPO := 823058932981.dkr.ecr.eu-west-1.amazonaws.com/pacioli
IMAGE_HASH := latest

pull:
	@echo "Pulling required images"
	docker pull 823058932981.dkr.ecr.eu-west-1.amazonaws.com/cockroach:latest-v21.1
	docker pull 823058932981.dkr.ecr.eu-west-1.amazonaws.com/tigerbeetle@sha256:b1fe98356a0db183b56b555eac17c5a43f4b61305f5ac711ea741d5085a2f977
	@echo "Pull complete"

testci: 
	@echo "Running tests"
	go test -p 2 ./...
	@echo "Tests complete."

test: testenvup
	@echo "Running tests"
	go test -p 2 ./...
	@echo "Not destroying test env as you may run tests again."
	@echo "Tests complete."

testenvup:
	@echo "Starting test environment"
	docker compose -f ./docker-compose-dev.yml up -d
	@echo "Test environment started"

testenvdown:
	@echo "Destroy test environment"
	docker compose -f ./docker-compose-dev.yml down
	@echo "Test environment destoyed"

format:
	@echo "Formatting"
	go fmt ./...
	@echo "Formate complete"

lint:
	@echo "Linting"
	golangci-lint run --go=1.17 ./...
	@echo "Lint complete"

build:
	@echo "Building image"
	DOCKER_BUILDKIT=1 docker build -t "$(IMAGE_REPO):$(IMAGE_HASH)" -f ./Dockerfile ./../
	@echo "Image built"

push:
	@echo "Pushing image"
	docker push "$(IMAGE_REPO):$(IMAGE_HASH)"
	@echo "Image pushed"
