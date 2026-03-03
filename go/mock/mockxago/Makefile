.PHONY: test unit-test e2e-test lint build clean help

help:
	@echo "MockXago Makefile targets:"
	@echo "  test         - Run linting and all tests (unit tests first, then e2e tests)"
	@echo "  lint         - Run golangci-lint"
	@echo "  unit-test    - Run only unit tests (with coverage report)"
	@echo "  e2e-test     - Run only e2e tests"
	@echo "  build        - Build the mockxago binary"
	@echo "  clean        - Remove build artifacts"

lint:
	@echo "Running golangci-lint..."
	golangci-lint run ./...

test: lint unit-test e2e-test
	@echo "All tests completed successfully"

unit-test:
	@echo "Running unit tests..."
	go test -count=1 -v -coverprofile=coverage.out ./internal/...
	@echo "\nCoverage Report:"
	@go tool cover -func=coverage.out | tail -1

e2e-test:
	@echo "Running e2e tests..."
	go test -count=1 -v -tags e2e ./testenv

build:
	go build -o mockxago ./cmd/mockxago

clean:
	rm -f mockxago
