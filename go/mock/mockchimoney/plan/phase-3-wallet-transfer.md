# Phase 3 — Wallet Transfer

**Goal**: The wallet-to-wallet transfer endpoint.

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

- [ ] Transfer between two existing sub-accounts succeeds
- [ ] Transfer requires `amountToSend`
- [ ] Transfer requires `originCurrency`
- [ ] Transfer requires `destinationCurrency`
- [ ] Transfer from non-existent sender returns 400
- [ ] The `sendViaInterledger` field is accepted and silently ignored
- [ ] All previous feature scenarios still pass
- [ ] All code passes `golangci-lint run ./...`
- [ ] Unit test coverage ≥ 50% (unit tests use memory-backed store)
- [ ] `make test` runs successfully (linting + unit tests + e2e tests)
