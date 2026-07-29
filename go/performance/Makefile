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
# COUNT wallets are split into SENDERS senders and the rest receivers.
COUNT ?= 20
SENDERS ?= 10
provision: build
	$(BIN) provision -count $(COUNT) -senders $(SENDERS) -fund -out $(OVERLAY)

test:
	go test ./...

lint:
	golangci-lint run ./...
