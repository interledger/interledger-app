BIN := ./bin/perf
CONFIG ?= scenarios/many-to-many.yaml
OVERLAY ?= wallets.local.yaml

.PHONY: build clean smoke run validate provision test lint

build:
	go build -o $(BIN) ./cmd/perf

clean:
	rm -rf ./bin

# Check a scenario without touching the system under test.
validate: build
	$(BIN) validate -config $(CONFIG) -config $(OVERLAY)

# Two payments per sender — verifies the harness end to end.
smoke: build
	$(BIN) run -config scenarios/smoke.yaml -config $(OVERLAY)

run: build
	$(BIN) run -config $(CONFIG) -config $(OVERLAY)

# Create local development wallets and write the overlay the targets above expect.
COUNTRIES ?= za,de,us
PER_COUNTRY ?= 20
TARGET ?= 5000
provision: build
	$(BIN) provision -countries $(COUNTRIES) -per-country $(PER_COUNTRY) -target $(TARGET) -out $(OVERLAY)

test:
	go test ./...

lint:
	golangci-lint run ./...
