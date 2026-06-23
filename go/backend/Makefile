
PG_HOST?=0.0.0.0

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

generate:
	@echo "Generating graphl"
	go generate ./...
	@echo "Graphql generated"

job:
	echo "package jobs" >> "./jobs/$$(date +%Y_%m)_$(name).go"

