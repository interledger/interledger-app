# Phase 3 — Wallet Transfer

**Goal**: The wallet-to-wallet transfer endpoint.

**Status**: Completed (2026-03-20)

**Definition of Done**: This phase is complete when `make test` passes successfully.

## Deliverables

- `internal/handler/transfer.go` — `POST /v0.2.4/multicurrency-wallets/transfer`
- Validate `amountToSend`, `originCurrency`, `destinationCurrency` as required
- Validate sender sub-account exists (if `subAccount` is provided)
- `sendViaInterledger` silently ignored

## Feature File

`features/wallet_management.feature` (transfer scenarios)

## Dependencies

- Phase 0 (bootstrap)
- Phase 1 (storage and wallet APIs)
- Phase 2 (authentication)

## Test-Driven Development Notes

1. **Red**: Run transfer scenarios in `wallet_management.feature`. Expect 404 or missing handler.
2. **Green**: Implement the handler with minimal validation. Return success for valid requests.
3. **Refactor**:
   - All three wallet handlers (`wallet.go`, `transfer.go`) should share validation helpers.
   - Deduplicate sub-account existence checks into a utility function.
   - Run all previous feature scenarios to ensure they still pass.

## Acceptance Criteria

- [x] Transfer between two existing sub-accounts succeeds
- [x] Transfer requires `amountToSend`
- [x] Transfer requires `originCurrency`
- [x] Transfer requires `destinationCurrency`
- [x] Transfer from non-existent sender returns 400
- [x] The `sendViaInterledger` field is accepted and silently ignored
- [x] All previous feature scenarios still pass
- [x] All code passes `golangci-lint run ./...`
- [x] Unit test coverage ≥ 50% (unit tests use memory-backed store)
- [x] `make test` runs successfully (linting + unit tests + e2e tests)

## Implementation Notes

- Added `POST /v0.2.4/multicurrency-wallets/transfer` in `internal/handler/transfer.go`.
- Added shared wallet validation helpers in `internal/handler/wallet_validation.go` and reused them from `wallet.go` and `transfer.go`.
- Added transfer route wiring in `cmd/mockchimoney/main.go`.
- Added transfer unit tests in `internal/handler/wallet_test.go` for all transfer feature scenarios.
- Added supporting coverage tests in `internal/config/config_test.go` and `internal/logger/logger_test.go` to keep total unit coverage above threshold.
- Validation run on 2026-03-20:
   - `make test`: pass
   - `golangci-lint run ./...`: pass
   - `go test ./... -coverprofile=coverage.out && go tool cover -func=coverage.out | tail -n 1`: `total: 63.2%`
