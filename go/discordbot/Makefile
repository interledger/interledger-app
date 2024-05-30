
IMAGE_REPO := 823058932981.dkr.ecr.eu-west-1.amazonaws.com/discordbot
IMAGE_HASH := latest
PG_HOST?=0.0.0.0

PF_SCRIPT=../backend/forward.sh

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

port-forward-crdb:
	@echo "Port forwarding cockroachdb"
	bash $(PF_SCRIPT) cockroachdb pod/cockroachdb-0 26257 26257

testci:
	@echo "Running tests"
	go test ./...
	@echo "Tests complete."

test:
	@echo "Running tests"
	make -C ../backend testenvup
	make -C ../backend generatetestmigrations
	go test ./...
	@echo "Tests complete."

