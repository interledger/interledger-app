export GOWORK=off

.PHONY: test lint

test: lint
	@echo "--- running unit tests ---"
	go test ./engine/... -coverprofile=coverage.out -coverpkg=./engine/...
	@go tool cover -func=coverage.out | tail -1
	@total=$$(go tool cover -func=coverage.out | awk '/^total:/ {gsub("%", "", $$3); print $$3}'); \
		echo "--- enforcing coverage >= 80% (current: $$total%) ---"; \
		awk -v c="$$total" 'BEGIN { if ((c + 0) < 80) { printf("coverage gate failed: %.1f%% < 80.0%%\n", c); exit 1 } }'

lint:
	@echo "--- running linter ---"
	golangci-lint run ./...

build:
	@echo "--- building TTT ---"
	go build -o ttt .

run:
	@echo "--- running TTT ---"
	go run .