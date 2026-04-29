export GOWORK=off

.PHONY: test lint e2e

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

e2e:
	@echo "--- running e2e tests ---"
	cd e2e && GOCACHE=/tmp/ttt-go-cache go test -count=1 ./... $(if $(or $(E2E_TAGS),$(E2E_DEBUG)),-args) $(if $(E2E_TAGS),-tags="$(E2E_TAGS)") $(if $(E2E_DEBUG),--debug="$(E2E_DEBUG)")

build:
	@echo "--- building TTT ---"
	go build -o ttt .

run:
	@echo "--- running TTT ---"
	go run .
