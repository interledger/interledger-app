# Phase 1: Foundations and Signup Core

## Objective
Establish the service foundation and deliver the minimum PTI API surface needed for signup and KYC initiation.

## Scope
- Project scaffold:
  - `cmd/mockpti`
  - `internal/*`
  - `testenv/*`
  - `Makefile`
  - `.golangci.yml`
- Health endpoint:
  - `GET /health`
- Core API endpoints:
  - `POST /users`
  - `GET /users/{id}`
  - `PATCH /users`
  - `PUT /users`
  - `POST /users/assessments`
  - `GET /users/{id}/assessments`
  - `POST /auth/jwt`
- Persistence seam from day one:
  - define store interface
  - memory store implementation
  - redis store implementation
  - dependency wiring through the interface

## TDD Rules
- For each endpoint behavior, first add failing unit tests.
- Implement the smallest code change that passes tests.
- Refactor immediately after green tests.

## Testing Requirements
- Unit tests use memory store only.
- E2E milestone coverage for Phase 1 feature files:
  - `features/service_health.feature`
  - `features/token_generation.feature`
  - `features/user_and_kyc.feature`
- Internal coverage gate: `go test -coverprofile=coverage.out ./internal/...` must report `>= 75.0%` total statements.

## Deliverables
- Running mock service with health and core signup/kyc endpoints.
- Deterministic validations and realistic status codes.
- Store interface with memory and redis implementations available.

## Acceptance Criteria
- `make lint`, `make unit-test`, `make build` pass.
- No domain code performs direct persistence outside the store interface.
- Endpoint behavior is covered with test-first unit tests.
- Phase 1 feature files pass in e2e.
- Internal package total coverage is `>= 75.0%`.

## Verification
Run `make test` from the `go/mock/mockpti` directory.
Run `go test -coverprofile=coverage.out ./internal/... && go tool cover -func=coverage.out | grep total` and verify total coverage is `>= 75.0%`.
The phase is complete only when both checks pass.
