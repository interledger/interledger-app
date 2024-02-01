
IMAGE_REPO := 823058932981.dkr.ecr.eu-west-1.amazonaws.com/backend
IMAGE_HASH := latest
CRDB_HOST?=0.0.0.0

PF_SCRIPT=./forward.sh

pull:
	@echo "Pulling required images"
	docker pull 823058932981.dkr.ecr.eu-west-1.amazonaws.com/cockroach:latest-v21.1
	docker pull 823058932981.dkr.ecr.eu-west-1.amazonaws.com/kratos:v0.8.2-alpha.1
	make -C ../pacioli build
	@echo "Pull complete"

testci:
	@echo "Running tests"
	go test ./...
	@echo "Tests complete."

generatetestmigrations:
	@echo "Generating test migrations"
	mkdir -p db/testmigrations
	rm -rf db/testmigrations/*
	atlas migrate diff create_all --dir "file://db/testmigrations" --to "file://db/schema.hcl" --dev-url="postgres://root@$(CRDB_HOST):26257?sslmode=disable"
	@echo "Done"

test: testenvup generatetestmigrations
	@echo "Running tests"
	go test ./... -p 4
	@echo "Tests complete."

testenvup:
	@echo "Starting test environment"
	docker compose -f ./docker-compose-dev.yml up -d
	sleep 2
	@echo "Test environment started"

testenvdown:
	@echo "Destroy test environment"
	docker compose -f ./docker-compose-dev.yml down -v
	@echo "Test environment destoyed"

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

generate:
	@echo "Generating graphl"
	go generate ./...
	@echo "Graphql generated"

port-forward-crdb:
	@echo "Port forwarding cockroachdb"
	bash $(PF_SCRIPT) cockroachdb pod/cockroachdb-0 26257 26257

port-forward-temporal-ui:
	@echo "Port forwarding temporal ui"
	bash $(PF_SCRIPT) temporal service/temporal-web 8080 8080

port-forward-rafiki:
	@echo "Port forwarding Rafiki admin"
	bash $(PF_SCRIPT) rafiki service/rafiki-rafiki-frontend 3010 3010

job:
	echo "package jobs" >> "./jobs/$$(date +%Y_%m)_$(name).go"

