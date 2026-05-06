# Phase 5 Implementation Plan: MockXago Unit Testing & Feature Import

## Overview

Phase 5 is the foundation phase for all subsequent Xago integration work. The goal is to establish a fully tested MockXago service with adequate code coverage and complete feature test coverage, using Redis as the backend storage.

**Key Outcome**: All unit and feature tests pass with 65%+ code coverage for `internal/` packages, and MockXago is ready for integration with the E2E test suite.

## Current Status Assessment

### Unit Test Coverage (Current: 44.0%)
- **auth**: 95.0% ✓ (excellent)
- **handler**: 24.0% ← **Critical gap** (needs expansion)
- **jobs**: 41.4% (needs expansion)
- **logger**: 0.0% (not tested)
- **models**: 0.0% (not tested)
- **storage**: 91.0% ✓ (excellent, memory implementation only)
- **utils**: 66.7% ✓ (acceptable)

**Target**: 65% overall coverage

### Feature Test Status
**Currently imported**: 5 features
- authentication.feature ✓
- balance_management.feature ✓
- beneficiary_management.feature ✓
- subaccount_management.feature ✓
- transactions.feature ✓

**Missing from bosbaber/mockxago**: 3 features
- currencies_and_deposits.feature ✗ (needs import)
- deposits_and_webhooks.feature ✗ (needs import)
- integration_workflows.feature ✗ (needs import)

### Storage Backend Status
- Currently: **In-memory only**
- Required: **Redis backend**
- Status: Redis implementation exists in `bosbaber/mockxago` as `internal/storage/redis.go` (757 lines)

## Phase 5 Tasks

### Task 0: Initial Setup - Local Environment (Sub-task of Phase 5 or Phase 6)
**Status**: Not started  
**Note**: These may be done as part of Phase 6 instead, but are listed here for completeness

This is a prerequisite for making MockXago accessible to E2E tests:

**Files to create**:
1. **Create `local/mockxago.yaml`** (from bosbaber/mockxago reference implementation)
   - MockXago service definition (Docker build, ports, environment)
   - TLS routing configuration (Traefik)
   - Redis connection settings
   - Webhook URL pointing to backend

2. **Update `local/wallet.yaml`** with XAGO environment variables:
   ```yaml
   - MOCKXAGO_ENDPOINT=http://mockxago:8080
   - XAGO_API_BASE_URL=http://mockxago:8080/v1
   - XAGO_IDENTITY_BASE_URL=http://mockxago:8080/v1
   - XAGO_API_PUBLIC_KEY=test-public-key
   - XAGO_API_SECRET=test-secret
   - XAGO_POLICY_ID=5e2585a474b0e90012ce8ff1
   - XAGO_USD_OPS_ACCOUNT=868196c3-f6b4-4920-bbfb-d1c7f6a98183
   - XAGO_ZAR_OPS_ACCOUNT=b0944908-16e6-4ef4-8677-192165e33c59
   - XAGO_LEDGER_ID_ZAR=9246927
   - XAGO_LEDGER_ID_USD=9246873
   ```

**Validation**:
- `local/mockxago.yaml` exists and is syntactically valid
- `local/wallet.yaml` contains all XAGO_* variables
- `docker-compose config` shows no errors

### Task 1: Import Missing Feature Files (1-2 hours)
**Status**: Not started

Bring over the three missing feature files from `bosbaber/mockxago` branch:

**Files to import**:
1. `go/mock/mockxago/features/currencies_and_deposits.feature`
2. `go/mock/mockxago/features/deposits_and_webhooks.feature`
3. `go/mock/mockxago/features/integration_workflows.feature`

**Action**:
```bash
git show bosbaber/mockxago:go/mock/mockxago/features/currencies_and_deposits.feature > \
  go/mock/mockxago/features/currencies_and_deposits.feature
git show bosbaber/mockxago:go/mock/mockxago/features/deposits_and_webhooks.feature > \
  go/mock/mockxago/features/deposits_and_webhooks.feature
git show bosbaber/mockxago:go/mock/mockxago/features/integration_workflows.feature > \
  go/mock/mockxago/features/integration_workflows.feature
```

**Validation**:
- All 8 feature files present: `ls go/mock/mockxago/features/*.feature`
- No duplicate step definitions in testenv

### Task 2: Import Missing Feature Files (1-2 hours)
**Status**: Not started

(Continuation from Task 1)

### Task 3: Implement Redis Storage Backend (2-3 hours)
**Status**: Not started

Port Redis storage implementation from `bosbaber/mockxago` to current branch.

**Files to create/modify**:
1. **New**: `go/mock/mockxago/internal/storage/redis.go` (from bosbaber/mockxago)
2. **New**: `go/mock/mockxago/internal/storage/redis_test.go` (if exists in bosbaber/mockxago)
3. **Modify**: `go/mock/mockxago/cmd/mockxago/main.go` (add Redis initialization logic)

**Implementation approach**:
- Copy complete Redis implementation from `bosbaber/mockxago:go/mock/mockxago/internal/storage/redis.go`
- Ensure Redis client configuration matches conventions used in mockgatehub:
  - Connection pooling (MinIdleConns, PoolSize)
  - Proper error handling and logging
  - Key namespace patterns (e.g., `token:`, `subaccount:`)
- Update main.go to support both memory and Redis:
  ```go
  var store storage.Storage
  redisURL := os.Getenv("MOCKXAGO_REDIS_URL")
  if redisURL != "" {
      store, err = storage.NewRedisStorage(redisURL, 0)
  } else {
      store = storage.NewMemoryStorage()
  }
  ```

**Validation**:
- `go build ./cmd/mockxago` succeeds
- Unit tests for redis storage pass: `go test ./internal/storage/... -v`
- Docker-compose can start mockxago with Redis (see Task 5)

### Task 3: Update Integration Test Infrastructure (1-2 hours)
**Status**: Not started

Configure mockxago to use Redis in integration tests.

**Files to modify**:
1. `go/mock/mockxago/testenv/docker-compose.yml`
   - Add Redis service (redis:7-alpine or similar)
   - Expose Redis on port 26379 or similar
   - Ensure mockxago container can reach Redis via `redis://redis:6379`
   - Set environment variable: `MOCKXAGO_REDIS_URL=redis://redis:6379`

2. `go/mock/mockxago/testenv/services.go`
   - Add startup health checks for Redis
   - Ensure Redis is cleaned between test runs (FLUSHDB)

**Example docker-compose addition**:
```yaml
services:
  redis:
    image: redis:7-alpine
    container_name: mockxago-redis-test
    ports:
      - "26379:6379"
    healthcheck:
      test: ["CMD", "redis-cli", "ping"]
      interval: 2s
      timeout: 1s
      retries: 5
    volumes:
      - mockxago-redis-data:/data

volumes:
  mockxago-redis-data:
```

**Validation**:
- `docker-compose up` in testenv succeeds
- Redis connectivity test: `docker exec mockxago-redis-test redis-cli ping`
- BDD tests run and connect to Redis without errors

### Task 5: Implement Handler Unit Tests (3-4 hours)
**Status**: Not started

Currently handler package has 24% coverage. Need targeted test expansion to reach target coverage. **DO THIS BEFORE fixing feature test steps.**

**Packages needing coverage expansion** (in priority order):

1. **handler package** (current: 24%) → Target: 60%+
   - `handler.go`: health checks, root handler, other utility handlers
   - `auth.go`: Create/Get/Update account endpoints
   - `beneficiary.go`: Beneficiary CRUD operations
   - `currency.go`: Currency listing and rates
   - `subaccount.go`: SubAccount creation, listing, KYC operations
   - Other handler files as needed

   **Test structure approach**:
   - Table-driven tests with testify assertions
   - Use existing `balance_test.go`, `currency_test.go`, `handler_test.go`, `subaccount_test.go` as templates
   - Mock storage interface where appropriate
   - Test both happy path and error cases

2. **models package** (current: 0%) → Target: Baseline coverage (any > 0%)
   - Add tests for model initialization, validations, marshaling/unmarshaling where applicable
   - Example: test struct field initialization, JSON marshaling, validation methods

3. **logger package** (current: 0%) → Target: Baseline coverage
   - Test logger initialization
   - Test log level configuration
   - Verify structured logging output format

4. **jobs package** (current: 41.4%) → Target: 75%+
   - Expand existing queue and job tests
   - Test job lifecycle (enqueue, process, retry, failure)
   - Test with Redis storage (add redis_test.go or expand existing)

**Implementation method**:
1. Run `go test ./internal/... -coverprofile=coverage.out` to establish baseline
2. For each package, run `go tool cover -html=coverage.out` to identify untested code paths
3. Add tests incrementally, validating coverage after each iteration
4. Ensure tests use both memory AND redis storage backends

**Validation**:
```bash
cd go/mock/mockxago
go test ./internal/... -coverprofile=coverage.out
go tool cover -func=coverage.out | grep total  # Must be >= 65%
```

### Task 6: Implement Missing Feature Step Definitions (2-3 hours)
**Status**: Not started

**IMPORTANT**: Only do this AFTER unit test coverage is ≥65%, to avoid context-switching.

The three new feature files will require step definitions in testenv/:
- `go/mock/mockxago/testenv/currencies_and_deposits_steps.go` (if needed)
- `go/mock/mockxago/testenv/deposits_and_webhooks_steps.go` (if needed)
- `go/mock/mockxago/testenv/integration_workflows_steps.go` (likely heavily needed from bosbaber/mockxago)

**Approach**:
1. Check existing step definitions across all `*_steps.go` files
2. Identify missing steps from new feature files: `go test ./testenv/ -v` will fail on undefined steps
3. Copy step implementations from `bosbaber/mockxago` testenv/ and adapt if necessary
4. Ensure all steps use Redis-backed storage (no hardcoded in-memory assumptions)

**Validation**:
```bash
cd go/mock/mockxago
go test -timeout 5m ./testenv/ -v  # All 8 feature files with all steps
```

### Task 7: Final Validation & Coverage Report (30 min)
**Status**: Not started

Run full test suite and generate coverage report.

**Actions**:
```bash
cd go/mock/mockxago

# Run all tests (unit + integration)
go test ./... -timeout 10m -coverprofile=coverage.out -v

# Generate coverage report
go tool cover -html=coverage.out -o coverage.html

# Display summary
go tool cover -func=coverage.out | tail -20

# Run BDD feature tests (if testenv infrastructure ready)
go test -timeout 10m ./testenv/... -v
```

**Exit criteria**:
- ✓ All unit tests pass
- ✓ Coverage >= 65% for `internal/...`
- ✓ All 8 feature files pass BDD tests
- ✓ Redis backend operational in docker-compose
- ✓ Coverage report generated at `go/mock/mockxago/coverage.html`

## Dependency Graph

```
Task 1: Import Feature Files
    ↓
Task 2: Redis Storage Backend
    ↓ (must complete before Task 3)
Task 3: Update Integration Infrastructure
    ↓ (runs in parallel with Task 4)
Task 4: Implement Handler Unit Tests ──→ Validation Step: 65%+ coverage achieved
    ↓
Task 5: Feature Step Definitions (only if Task 4 ✓)
    ↓
Task 6: Final Validation & Coverage
```

## Testing Before Proceeding to Phase 6

**CRITICAL**: The following validation MUST PASS before Phase 6 can proceed:

```bash
cd /home/stephan/interledger/interledger-app/go/mock/mockxago

# 1. Unit test coverage check
go test ./internal/... -coverprofile=coverage.out -v
go tool cover -func=coverage.out | grep "total.*%" 
# Must show: total >= 65%

# 2. All feature tests pass
go test ./testenv/... -v -timeout 10m
# Must show: 0 failures

# 3. No linting issues
go vet ./...  # or golangci-lint if configured

# 4. Binary builds successfully
go build ./cmd/mockxago

# 5. Docker image builds successfully
docker build -f Dockerfile -t mockxago:test .
```

**Failed test artifact preservation**:
- If any tests fail, preserve:
  - Coverage report: `go/mock/mockxago/coverage.out`
  - Test output: `go test ./... -v > test-output.log`
  - Screenshots (if any GH action): `e2e/debug/`


## Success Criteria

Phase 5 is **COMPLETE** when:

1. ✅ All 8 feature files present in `go/mock/mockxago/features/`
2. ✅ Redis storage implementation present in `go/mock/mockxago/internal/storage/redis.go`
3. ✅ Redis service configured in `testenv/docker-compose.yml`
4. ✅ Unit test coverage >= 65% for `internal/...` packages
5. ✅ All 8 feature files pass `go test ./testenv/... -v`
6. ✅ No linting errors: `go vet ./...`
7. ✅ Binary builds successfully: `go build ./cmd/mockxago`
8. ✅ Docker image builds: `docker build -f Dockerfile -t mockxago:test .`

**All tests must pass before Phase 6 can proceed.**

## Risk Mitigation

| Risk | Mitigation |
|------|-----------|
| Redis connection issues in docker-compose | Test redis CLI commands during Task 3, implement health checks |
| Feature tests require new steps not in bosbaber/mockxago | Document undefined steps during test runs, implement based on feature requirements |
| Handler test expansion takes longer than estimated | Prioritize high-impact areas (auth, subaccount, balance) first |
| Code coverage measurement differences | Use `go tool cover` as source of truth, consistent with CI/CD pipeline |

## Notes for Implementation Team

- **Do not** attempt to fix feature test steps until unit test coverage is >= 65%. This avoids context-switching.
- **Focus on coverage** in Task 4 — coverage == quality assurance.
- **Redis is mandatory** — all subsequent phases depend on Redis being operational.
- **Commit frequently** after each task to maintain audit trail and enable rollback if needed.
- Use `git show bosbaber/mockxago:path/to/file` to reference existing code rather than checking out the branch.

---

**Last Updated**: March 6, 2026
**Phase**: 5 (Foundation)
**Next Phase**: Phase 6 (Backend Integration)
