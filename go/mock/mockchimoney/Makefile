.PHONY: help test unit-tests e2e-tests coverage build lint clean run docker-build

REDIS_TEST_CONTAINER ?= mockchimoney-e2e-redis
REDIS_TEST_PORT ?= 26381
REDIS_TEST_DB ?= 5

help:
	@echo "MockChimoney Commands"
	@echo ""
	@echo "build             Build the mockchimoney binary"
	@echo "run               Run the service"
	@echo "test              Run unit tests + feature e2e tests"
	@echo "unit-tests        Run unit tests only"
	@echo "e2e-tests         Run feature (godog) e2e tests"
	@echo "coverage          Run unit tests with coverage report"
	@echo "lint              Run linter (golangci-lint)"
	@echo "docker-build      Build Docker image"
	@echo "clean             Clean up build artifacts"
	@echo ""

# Run all tests: linting + unit tests + feature tests
test: lint unit-tests e2e-tests
	@echo ""
	@echo "✅ All tests completed"

# Run unit tests
unit-tests:
	@echo "Running all unit tests..."
	@go test -v ./...
	@echo "Checking coverage for internal"
	@go test -v -coverprofile=coverage.out ./internal/...

# Run feature e2e tests
e2e-tests:
	@if [ -d testenv ]; then \
		echo "Running feature e2e tests (godog) with Redis backend..."; \
		set -e; \
		cleanup() { docker rm -f $(REDIS_TEST_CONTAINER) >/dev/null 2>&1 || true; }; \
		trap cleanup EXIT; \
		docker rm -f $(REDIS_TEST_CONTAINER) >/dev/null 2>&1 || true; \
		docker run -d --name $(REDIS_TEST_CONTAINER) -p $(REDIS_TEST_PORT):6379 valkey/valkey:8-alpine >/dev/null; \
		MOCKCHIMONEY_REDIS_URL=redis://localhost:$(REDIS_TEST_PORT) MOCKCHIMONEY_REDIS_DB=$(REDIS_TEST_DB) go test -tags e2e -v -count=1 ./testenv/ -run TestFeatures; \
	else \
		echo "Skipping e2e tests (testenv not yet created)"; \
	fi

# Run tests with coverage report
coverage:
	@echo "Running unit tests with coverage..."
	@go test -v ./... -coverprofile=coverage.out
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Build the mockchimoney binary
build:
	@echo "Building mockchimoney..."
	@go build -v -o mockchimoney ./cmd/mockchimoney

# Run the service
run: build
	@echo "Running mockchimoney..."
	@./mockchimoney

# Run linter
lint:
	@echo "Running linters..."
	@echo "Running golangci-lint..."
	golangci-lint run ./...

# Build Docker image
docker-build:
	@echo "Building Docker image for mockchimoney..."
	docker build -t mockchimoney:latest .

# Clean up
clean:
	@echo "Cleaning up..."
	@rm -f mockchimoney coverage.out coverage.html
	@go clean -testcache
	@echo "Clean complete"
