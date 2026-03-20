# Phase 3: Transactions

## Objective
Deliver PTI transaction endpoints with deterministic, testable status progression.

## Scope
- Transaction endpoints:
  - `POST /transactions/deposits`
  - `POST /transactions/withdrawals`
  - `POST /transactions/transfers`
  - `GET /transactions/{requestId}`
  - `POST /transactions/{requestId}/updates`
- Deterministic progression logic:
  - default: `PENDING -> SETTLED`
  - configurable failure path for test scenarios

## TDD Rules
- Each transaction behavior starts with failing unit tests.
- Refactor after green tests to keep code simple and composable.

## Persistence and Architecture Rules
- Transaction state is persisted via the shared store interface only.
- Unit tests use memory store.
- E2E scenarios use redis store.

## Testing Requirements
- Unit coverage for progression and error/failure paths.
- E2E coverage for deposit/withdraw/transfer creation and status reads.
- Feature-file milestone for this phase:
  - `features/transactions.feature`
- Internal coverage gate: `go test -coverprofile=coverage.out ./internal/...` must report `>= 75.0%` total statements.

## Deliverables
- Deterministic transaction APIs suitable for backend integration tests.
- Scenario controls sufficient for predictable negative-path testing.

## Acceptance Criteria
- Transaction endpoints pass unit/e2e coverage for happy and failure paths.
- Status transitions are deterministic and documented.
- Phase 1 and Phase 2 suites remain green.
- `features/transactions.feature` passes end-to-end.
- Internal package total coverage is `>= 75.0%`.

## Verification
Run `make test` from the `go/mock/mockpti` directory.
Run `go test -coverprofile=coverage.out ./internal/... && go tool cover -func=coverage.out | grep total` and verify total coverage is `>= 75.0%`.
The phase is complete only when both checks pass.
