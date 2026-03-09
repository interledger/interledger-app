.PHONY: test unit-test e2e-test build clean help

help:
	@echo "MockXago Makefile targets:"
	@echo "  test         - Run all tests (unit tests first, then e2e tests)"
	@echo "  unit-test    - Run only unit tests"
	@echo "  e2e-test     - Run only e2e tests"
	@echo "  build        - Build the mockxago binary"
	@echo "  clean        - Remove build artifacts"

test: unit-test e2e-test
	@echo "All tests completed successfully"

unit-test:
	@echo "Running unit tests..."
	go test -count=1 -v ./cmd/... ./internal/...
	@echo "Checking coverage for internal"
	@go test -v -coverprofile=coverage.out ./internal/...
	@go tool cover -html=coverage.out -o coverage.html

e2e-test:
	@echo "Running e2e tests..."
	go test -count=1 -v -tags e2e ./testenv

build:
	go build -o mockxago ./cmd/mockxago

clean:
	rm -f mockxago
