.PHONY: test lint

test: lint
	@echo "--- running unit tests ---"
	go test ./engine/... -coverprofile=coverage.out -coverpkg=./engine/...
	@go tool cover -func=coverage.out | tail -1

lint:
	@echo "--- running linter ---"
	golangci-lint run ./...
