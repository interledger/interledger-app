# MegaAccounts

MegaAccounts is a terminal-based ledger simulator for multi-provider wallet flows.
It models double-entry bookkeeping across providers, currencies, users, liquidity pools, and bilateral settlement.

The app is built with a strict split between:
- `engine`: accounting domain logic and invariants
- `gui`: Bubble Tea TUI for running workflows and inspecting state

## Features

### Ledger and accounting model
- Pure double-entry ledger with per-event posting.
- Integer-only amounts (no floating-point ledger math).
- Credit-normal balance semantics (`credit = +`, `debit = -`).
- Currency-aware global balancing checks.

### Account types
- **System account** per provider (printer/origin account).
- **Liquidity account** per provider and currency.
- **Position account** per liquidity account and counterparty provider.
- **User account** per user + provider + currency.

### Implemented workflows
- **Fund Provider Liquidity**
  - Debits provider system account and credits provider liquidity.
- **User Onboard**
  - Debits provider system account and credits user account.
- **P2P Transfer (Same Provider)**
  - Debits sender and credits recipient.
- **User Offboard**
  - Debits user and credits provider system account.
- **Cross-Provider Transfer**
  - Moves value from sender user/provider/currency to recipient user/provider/currency.
  - Automatically creates missing position accounts.
  - Stores FX metadata on each ledger entry.
  - Uses internal FX service (no manual rate input in UI).
- **Bilateral Settlement**
  - Settles net position between two providers for a currency up to a cutoff time.
- **Clear Everything (DANGER)**
  - Wipes providers/accounts/entries from the store after explicit confirmation (`CLEAR`).
  - Re-seeds default demo providers/accounts.

### Variable FX simulation (Phase 4)
- Internal FX service with integer rational rates (`num/den`).
- Starts with `1 EUR = 15 ZAR` at app boot.
- Each successful cross-provider conversion mutates the rate by exactly `+5%` or `-5%`.
- Mutation direction is random in production and injectable/deterministic for tests.
- Failed conversions do **not** mutate the FX rate.
- Historical event metadata preserves the exact rate used at execution time.

### Integrity checks
Main view continuously runs and displays:
- Per-event zero-sum check (last event)
- Global per-currency zero-sum check
- Liquidity decomposition check (`float = liquidity - sum(position balances)`)
- Bilateral mirror check (position pairs must sum to zero)

### TUI capabilities
- Main ledger table with horizontal and vertical navigation.
- Frozen metadata columns (`Time`, `Workflow`, `Step`).
- Wide step descriptions including FX rate marker in step text.
- **Balances panel** at the top:
  - Shows account balances at the point in time of the currently highlighted row.
  - Includes event/workflow context for that highlighted entry.
- Live FX section showing current internal FX quote(s).
- Workflow menu with guided forms and dynamic user dropdowns where applicable.

### Storage
- Default runtime store is SQLite via `modernc.org/sqlite` (pure Go; no CGO required).
- DB path defaults to `megaaccounts.db` in the working directory.
- Override with environment variable:
  - `MEGAACCOUNTS_DB=/path/to/file.db`
- In-memory store implementation also exists for tests/dev.

### Test coverage
- Extensive engine tests for workflows, checks, FX behavior, and store parity.
- SQLite and memory stores validated for equivalent behavior in scenario tests.

## Project layout

- `main.go`: app bootstrap, SQLite wiring, FX initialization, seed data
- `engine/`: domain model, workflows, checks, FX service, store interface
- `engine/memory/`: in-memory store backend
- `engine/sqlite/`: SQLite store backend and persistence tests
- `gui/`: Bubble Tea model/update/view and workflow forms
- `specs/`: phase docs and architecture/requirements notes

## Requirements

- Go `1.25+`

Optional for `make test`:
- `golangci-lint` installed and on PATH

## Getting started

```bash
go mod tidy
go run .
```

## Build

```bash
go build ./...
```

## Test

Run unit tests directly:

```bash
go test ./...
```

Run lint + coverage workflow from Makefile:

```bash
make test
```

## Cross-compile for macOS

Apple Silicon:

```bash
CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -o dist/megaaccounts-darwin-arm64 .
```

Intel macOS:

```bash
CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -o dist/megaaccounts-darwin-amd64 .
```

## Notes for simulation use

- Amounts are entered as decimal values in UI forms and converted to base units internally.
- On startup, demo entities are seeded idempotently:
  - Providers: `gatehub`, `xago`
  - Accounts: system/liquidity combos for EUR/ZAR as configured
  - Users:
    - `alice` on `gatehub` (`EUR`)
    - `bob` on `gatehub` (`EUR`)
    - `carlos` on `xago` (`ZAR`)
- Because seeding is idempotent, restarting the app preserves existing DB data.

## License

No license file is currently defined in this repository.
