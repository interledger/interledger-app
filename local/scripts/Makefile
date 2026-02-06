.PHONY: build test coverage clean run install help

# Binary name
BINARY=local-dev-tool

# Build the application
build:
	@echo "Building ${BINARY}..."
	go build -o ${BINARY} .
	@echo "✓ Build complete: ./${BINARY}"

# Run tests
test:
	@echo "Running tests..."
	go test -v ./...

# Generate coverage report
coverage:
	@echo "Generating coverage report..."
	go test -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out
	@echo "\nTo view HTML coverage report, run:"
	@echo "  go tool cover -html=coverage.out"

# Generate and open HTML coverage report
coverage-html: coverage
	go tool cover -html=coverage.out

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	rm -f ${BINARY}
	rm -f coverage.out
	@echo "✓ Clean complete"

# Run the application (Rafiki setup)
run-rafiki:
	go run main.go rafiki

# Run with all assets (no UI)
run-rafiki-all:
	go run main.go rafiki --skip-ui

# Install dependencies
install:
	@echo "Installing dependencies..."
	go mod download
	go mod tidy
	@echo "✓ Dependencies installed"

# Format code
fmt:
	@echo "Formatting code..."
	go fmt ./...
	@echo "✓ Code formatted"

# Run linter (requires golangci-lint)
lint:
	@echo "Running linter..."
	golangci-lint run ./...

# Quick build and test
all: install fmt test build

# Help
help:
	@echo "Local Development Tool - Makefile"
	@echo ""
	@echo "Usage: make <target>"
	@echo ""
	@echo "Targets:"
	@echo "  build           Build the application binary"
	@echo "  test            Run all tests"
	@echo "  coverage        Generate test coverage report"
	@echo "  coverage-html   Generate and open HTML coverage report"
	@echo "  clean           Remove build artifacts"
	@echo "  run-rafiki      Run Rafiki setup (interactive)"
	@echo "  run-rafiki-all  Run Rafiki setup (all assets, no UI)"
	@echo "  install         Install dependencies"
	@echo "  fmt             Format code"
	@echo "  lint            Run linter (requires golangci-lint)"
	@echo "  all             Install, format, test, and build"
	@echo "  help            Show this help message"
