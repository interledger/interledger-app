.PHONY: help test unit-tests e2e-tests coverage build lint clean

help:
	@echo "MockGatehub Test Commands"
	@echo ""
	@echo "test              Run unit tests + feature e2e tests"
	@echo "unit-tests        Run unit tests only"
	@echo "e2e-tests         Run feature (godog) e2e tests"
	@echo "coverage          Run unit tests with coverage report"
	@echo "build             Build the mockgatehub binary"
	@echo "lint              Run linter (gofmt, go vet)"
	@echo "clean             Clean up build artifacts and test binaries"
	@echo ""

# Run all tests: unit tests + feature tests
test: lint unit-tests e2e-tests
	@echo ""
	@echo "✅ All tests completed"

# Run unit tests
unit-tests:
	@echo "Running all unit tests..."
	@go test -v ./...
	@echo "Checking coverage for internal"
	@go test -v -coverprofile=coverage.out ./internal/...
	

e2e-tests:
	@echo "Running feature e2e tests (godog)..."
	@go test -tags e2e -v -count=1 ./testenv/ -run TestFeatures

# Run tests with coverage report
coverage:
	@echo "Running unit tests with coverage..."
	@go test -v ./... -coverprofile=coverage.out
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Build the mockgatehub binary
build:
	@echo "Building mockgatehub..."
	@go build -v -o mockgatehub ./cmd/mockgatehub

# Run linter
lint:
	@echo "Running linters..."
	@gofmt -l .
	@go vet ./...

# Clean up artifacts
clean:
	@echo "Cleaning up..."
	@rm -f mockgatehub coverage.out coverage.html
	@cd testenv && rm -f test_e2e
	@go clean -testcache
	@echo "Clean complete"
