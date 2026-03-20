# Phase 2 — Authentication Middleware

**Goal**: `X-API-KEY` validation on all non-health endpoints, controlled by an env var toggle.

## Deliverables

- `internal/handler/middleware.go` — `APIKeyMiddleware(key string, enforce bool) func(http.Handler) http.Handler`
- Apply middleware in `main.go` after the health route but before all API routes
- Unit tests for the middleware in `internal/handler/middleware_test.go`

## Feature File

`features/authentication.feature`

## Dependencies

- Phase 0 (bootstrap) must be complete
- Phase 1 (wallet APIs) must be complete and green

## Test-Driven Development Notes

1. **Red**: Run feature scenarios. Expect `POST /v0.2.4/multicurrency-wallets/create` to succeed even without the header (breaking change from Phase 1).
2. **Green**: Implement the middleware. Wire it into the router before all protected routes. Health route must be excluded.
3. **Refactor**:
   - Ensure `MOCKCHIMONEY_ENFORCE_AUTHENTICATION=false` path works end-to-end.
   - Unit test the middleware in isolation (mock request context, etc.).
   - Run all previous feature scenarios to ensure they still pass.

## Acceptance Criteria

- [ ] Valid API key is accepted
- [ ] Missing X-API-KEY header is rejected with 401 when enforced
- [ ] Wrong API key is rejected with 401 when enforced
- [ ] All protected endpoints (`payment`, `payout`, `wallet`, etc.) require the key
- [ ] Health check does not require authentication even when enforced
- [ ] Authentication can be disabled for development (`MOCKCHIMONEY_ENFORCE_AUTHENTICATION=false`)
- [ ] All previous feature scenarios still pass
- [ ] All code passes `golangci-lint run ./...`- [ ] Unit test coverage ≥ 50% (unit tests use memory-backed store)
- [ ] `make test` runs successfully (linting + unit tests + e2e tests)