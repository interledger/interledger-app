.PHONY: test unit-test e2e-test build clean help lint

help:
	@echo "MockPTI Makefile targets:"
	@echo "  test         - Run all tests (lint, unit tests, then e2e tests)"
	@echo "  unit-test    - Run only unit tests"
	@echo "  e2e-test     - Run only e2e tests"
	@echo "  lint         - Run linters"
	@echo "  build        - Build the mockpti binary"
	@echo "  clean        - Remove build artifacts"

test: lint unit-test e2e-test
	@echo "All tests completed successfully"

unit-test:
	@echo "Running unit tests..."
	go test -count=1 -v ./cmd/... ./internal/...
	@echo "Checking coverage for internal"
	@go test -v -coverprofile=coverage.out ./internal/...
	@go tool cover -html=coverage.out -o coverage.html

lint:
	@echo "Running linters..."
	@gofmt -l .
	@go vet ./...
	@echo "Running golangci-lint..."
	golangci-lint run ./...

e2e-test:
	@echo "Running e2e tests..."
	go test -count=1 -v -tags e2e ./testenv

build:
	go build -o mockpti ./cmd/mockpti

clean:
	rm -f mockpti coverage.out coverage.html
