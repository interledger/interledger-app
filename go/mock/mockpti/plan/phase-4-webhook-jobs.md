# Phase 4: Webhook Jobs and Delivery

## Objective
Implement asynchronous webhook delivery via a persisted jobs mechanism compatible with backend `/webhooks/pti` handling.

## Scope
- Outgoing webhook jobs model and queue.
- Async worker(s) that process queued jobs.
- Webhook payload emission for:
  - `USER_ASSESSMENT`
  - `TRANSACTION_STATUS`
- Delivery mode support:
  - `plain` payload mode for local/initial milestone
  - `signed/encrypted` compatibility mode as follow-up enhancement

## Mandatory Queue Rule
- Outgoing webhook HTTP calls must be produced and delivered through persisted jobs.
- Direct synchronous HTTP delivery from domain handlers is not allowed.

## Job Persistence Requirements
- Persist job lifecycle states (minimum):
  - `queued`
  - `processing`
  - `delivered`
  - `failed` (retryable/terminal semantics)
- Persist retries and backoff metadata through the shared store interface.

## TDD Rules
- Test-first for enqueue, worker processing, retry, and terminal failure behavior.
- Refactor after each behavior lands.

## Testing Requirements
- Unit tests on memory store for queue semantics.
- E2E tests on redis store for end-to-end delivery and persistence semantics.
- Feature-file milestone for this phase:
  - `features/webhooks.feature`
- Internal coverage gate: `go test -coverprofile=coverage.out ./internal/...` must report `>= 65.0%` total statements.

## Deliverables
- Reliable webhook delivery subsystem with observable persisted state transitions.
- Backend-compatible webhook emission for signup and transactions.

## Acceptance Criteria
- Queue behavior is fully tested and deterministic.
- Webhook delivery is resilient to transient failures (retry behavior verified).
- By phase end, backend signup flow can complete through webhook updates.
- `features/webhooks.feature` passes end-to-end.
- Internal package total coverage is `>= 65.0%`.

## Verification
Run `make test` from the `go/mock/mockpti` directory.
Run `go test -coverprofile=coverage.out ./internal/... && go tool cover -func=coverage.out | grep total` and verify total coverage is `>= 75.0%`.
The phase is complete only when both checks pass.
