# Phase 9 — Redis Storage Backend

**Goal**: Persistent storage that survives service restarts, using the same Redis instance as the rest of the local environment.

## Deliverables

- `internal/storage/redis.go` — Redis implementation of `Store` interface, using `MOCKCHIMONEY_REDIS_URL` and `MOCKCHIMONEY_REDIS_DB`
- Wire in `main.go`: if `MOCKCHIMONEY_REDIS_URL` is non-empty, use Redis; otherwise use in-memory
- `testenv/docker-compose.yml` — adds a Redis service for BDD integration tests (if not already present via the main environment)

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

- [ ] `MOCKCHIMONEY_REDIS_URL` env var activates Redis storage
- [ ] Empty `MOCKCHIMONEY_REDIS_URL` falls back to in-memory
- [ ] `MOCKCHIMONEY_REDIS_DB` selects the Redis database number
- [ ] All feature scenarios pass with Redis backend
- [ ] Sub-account creation/retrieval survives service restart (manual test)
- [ ] Payment records persist across service restarts
- [ ] Payout records persist across service restarts
- [ ] All previous feature scenarios still pass
- [ ] All code passes `golangci-lint run ./...`
- [ ] Unit test coverage ≥ 60% (unit tests use memory-backed store); E2E tests use redis-backed store
- [ ] `make test` runs successfully (linting + unit tests + e2e tests)

## Notes

- Redis key design should avoid collisions with other services in the environment (use the `chimoney:` prefix).
- Consider TTL (time-to-live) for webhook job queue entries if using Redis for the queue as well (future enhancement).
