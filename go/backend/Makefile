
IMAGE_REPO := 823058932981.dkr.ecr.eu-west-1.amazonaws.com/backend
IMAGE_HASH := latest
PG_HOST?=0.0.0.0

pull:
	@echo "Pulling required images"
	docker pull 823058932981.dkr.ecr.eu-west-1.amazonaws.com/cockroach:latest-v21.1
	docker pull 823058932981.dkr.ecr.eu-west-1.amazonaws.com/kratos:v0.8.2-alpha.1
	make -C ../pacioli build
	@echo "Pull complete"

generatetestmigrations:
	@echo "Generating test migrations"
	mkdir -p db/testmigrations
	rm -rf db/testmigrations/*
	atlas migrate diff create_all --dir "file://db/testmigrations" --to "file://db/schema.hcl" --dev-url="postgres://postgres:password@$(PG_HOST):5432?sslmode=disable"
	@echo "Done"

test: testenvup generatetestmigrations
	@echo "Running tests"
	go test -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out

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

port-forward-db-dev:
	@echo "Port forwarding dev db"
	cloudflared access tcp --hostname hetzner.mgnt.fynbos.dev/db-dev --url localhost:6432

port-forward-db-prod:
	@echo "Port forwarding prod db"
	cloudflared access tcp --hostname hetzner.mgnt.fynbos.dev/db-prod --url localhost:5432

port-forward-temporal-dev:
	@echo "Port forwarding dev temporal"
	cloudflared access tcp --hostname hetzner.mgnt.fynbos.dev/temporal-dev --url localhost:7233

port-forward-temporal-prod:
	@echo "Port forwarding prod temporal"
	cloudflared access tcp --hostname hetzner.mgnt.fynbos.dev/temporal-prod --url localhost:7233

job:
	echo "package jobs" >> "./jobs/$$(date +%Y_%m)_$(name).go"

devdeploy:
	@echo "Deploying to local environment..."
	(cd ../../dev/vagrant && vagrant ssh -c 'nomad job run /home/vagrant/fynbos/dev/nomad/backend.hcl')

