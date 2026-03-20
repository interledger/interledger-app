# Phase 9 — Redis Storage Backend

**Goal**: Persistent storage that survives service restarts, using the same Redis instance as the rest of the local environment.

**Definition of Done**: This phase is complete when `make test` passes successfully.

## Deliverables

- `internal/storage/redis.go` — Redis implementation of `Store` interface, using `MOCKCHIMONEY_REDIS_URL` and `MOCKCHIMONEY_REDIS_DB`
- Wire in `main.go`: if `MOCKCHIMONEY_REDIS_URL` is non-empty, use Redis; otherwise use in-memory
- `testenv/docker-compose.yml` — adds a Redis service for BDD integration tests (if not already present via the main environment)

## Implementation Notes (2026-03-20)

- Added `internal/storage/redis.go` with a Redis-backed `Store` implementation:
   - key namespace prefix: `chimoney:`
   - keys used: `chimoney:subaccount:{id}`, `chimoney:payment:{issueID}`, `chimoney:payout:{issueID}`, `chimoney:payout:chiref:{chiRef}`
   - sub-account index set: `chimoney:subaccounts`
   - `SetNX` used for create operations to preserve `ErrAlreadyExists` semantics
   - `redis.Nil` mapped to `ErrNotFound`
   - `Close()` and `FlushAll()` helper methods added for lifecycle/test support
- Updated `internal/config/config.go` and tests:
   - added `RedisURL` (`MOCKCHIMONEY_REDIS_URL`, default `""`)
   - added `RedisDB` (`MOCKCHIMONEY_REDIS_DB`, default `5`)
- Updated `cmd/mockchimoney/main.go`:
   - selects Redis store when `MOCKCHIMONEY_REDIS_URL` is non-empty
   - falls back to in-memory store when empty
   - logs backend configuration and closes Redis client on shutdown
- Added shared storage contract tests:
   - `internal/storage/store_contract_test.go`
   - `internal/storage/memory_test.go` now runs contract suite
   - `internal/storage/redis_test.go` runs same suite with `miniredis`
   - added persistence test across store instances (simulated restart) for sub-accounts, payments, and payouts
- Added `testenv/docker-compose.yml` for Redis-backed BDD integration environment
- Updated `testenv/context.go`:
   - store type changed to `storage.Store`
   - supports Redis backend via `MOCKCHIMONEY_REDIS_URL` / `MOCKCHIMONEY_REDIS_DB`
   - flushes Redis DB at scenario bootstrap to avoid cross-scenario leakage while preserving data on explicit in-scenario restarts

## Validation Run

- `make test` (lint + unit + e2e): passed
- Redis-backed feature run:
   - `MOCKCHIMONEY_REDIS_URL=redis://localhost:26381 MOCKCHIMONEY_REDIS_DB=5 go test -tags e2e -count=1 ./testenv -run TestFeatures`
   - passed
- Coverage check:
   - `go test ./... -coverprofile=coverage.out`
   - `go tool cover -func=coverage.out`
   - total coverage: `68.1%`

## Feature Files

All previous feature files should continue to pass with Redis as the backend.

## Dependencies

- Phases 0–8 must be complete and green
- Redis client library (e.g., `github.com/redis/go-redis/v9`)

## Test-Driven Development Notes

1. **Red**: Ensure all feature scenarios pass with in-memory storage (from previous phases).
2. **Green**:
   - Implement the `Store` interface using Redis commands.
   - Wire the factory function in `main.go` to select storage backend.
   - Use Redis keys like `chimoney:subaccount:{id}`, `chimoney:payment:{issueID}`, etc.
3. **Refactor**:
   - The `Store` interface should not change from previous phases.
   - Consider a shared test suite that runs both `memory.go` and `redis.go` implementations against the same assertions (parametrized tests).
   - Ensure connection pooling and graceful shutdown.
   - Run all previous feature scenarios with both backends.

## Acceptance Criteria

- [x] `MOCKCHIMONEY_REDIS_URL` env var activates Redis storage
- [x] Empty `MOCKCHIMONEY_REDIS_URL` falls back to in-memory
- [x] `MOCKCHIMONEY_REDIS_DB` selects the Redis database number
- [x] All feature scenarios pass with Redis backend
- [x] Sub-account creation/retrieval survives service restart (validated via Redis store persistence test across store instances)
- [x] Payment records persist across service restarts (validated via Redis store persistence test across store instances)
- [x] Payout records persist across service restarts (validated via Redis store persistence test across store instances)
- [x] All previous feature scenarios still pass
- [x] All code passes `golangci-lint run ./...`
- [x] Unit test coverage ≥ 60% (unit tests use memory-backed store); E2E tests use redis-backed store
- [x] `make test` runs successfully (linting + unit tests + e2e tests)

## Notes

- Redis key design should avoid collisions with other services in the environment (use the `chimoney:` prefix).
- Consider TTL (time-to-live) for webhook job queue entries if using Redis for the queue as well (future enhancement).
