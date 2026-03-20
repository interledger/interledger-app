# Phase 1 — Storage Layer and Wallet APIs

**Goal**: Persistent (in-memory) storage for sub-accounts, plus the create/get wallet endpoints.

**Definition of Done**: This phase is complete when `make test` passes successfully.

## Deliverables

- `internal/models/models.go` — `SubAccount`, `Payment`, `Payout` structs (Go-idiomatic, no JSON tags yet)
- `internal/models/api.go` — `APIResponse` envelope: `{"status":"success"|"error","error":"...","data":{...}}`
- `internal/storage/interface.go` — `Store` interface: `CreateSubAccount`, `GetSubAccount`, `ListSubAccounts`
- `internal/storage/memory.go` — goroutine-safe in-memory map implementation
- `internal/handler/wallet.go` — `POST /v0.2.4/multicurrency-wallets/create`, `GET /v0.2.4/multicurrency-wallets/get`

## Feature File

`features/wallet_management.feature` (create and get scenarios only; transfer scenarios in Phase 3)

## Dependencies

- Phase 0 (bootstrap) must be complete and green

## Test-Driven Development Notes

1. **Red**: Run feature scenarios for wallet creation and retrieval. Expect 404s and missing handler errors.
2. **Green**: Implement the handlers using the in-memory store. Return proper JSON response envelopes.
3. **Refactor**:
   - The response shaping (wrapping data in `APIResponse`) should be a generic helper, not repeated per handler.
   - Extract `respondOK(w, data)` / `respondErr(w, statusCode, msg)` helpers into a utilities file.
   - Ensure all previous feature scenarios (from Phase 0) still pass.

## Acceptance Criteria

- [ ] Wallet creation with only `name` field succeeds
- [ ] Wallet creation with all optional fields (`email`, `firstName`, `lastName`, `phoneNumber`) succeeds
- [ ] Wallet creation without `name` returns 400
- [ ] Each wallet receives a unique ID
- [ ] Newly created wallet has `verification.status = "pending"`
- [ ] `GET /multicurrency-wallets/get?id=<id>` retrieves a wallet by ID
- [ ] Unknown wallet ID returns 404
- [ ] Missing `id` query param returns 400
- [ ] All previous feature scenarios still pass
- [ ] All code passes `golangci-lint run ./...`- [ ] `make test` runs successfully (linting + unit tests + e2e tests)