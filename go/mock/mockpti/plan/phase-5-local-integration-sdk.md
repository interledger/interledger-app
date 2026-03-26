# Phase 5: Local Integration and SDK Reliability

## Objective
Make mock PTI first-class in local development and stabilize browser e2e behavior for signup-related flows.

## Scope
- Local compose integration:
  - add `local/mockpti.yaml`
  - include in `local/docker-compose.yaml`
  - backend dependency in `local/wallet.yaml`
- Local defaults and env switches:
  - `PTI_BASE_URL` local default to mockpti
  - keep real PTI override path documented in `.env`
- Local host/cert updates:
  - `mockpti.interledger.test` in hosts/SAN/readme
- Backend widget configurability:
  - env-driven `PTI_SDK_URL` and `PTI_FORMS_URL`
- Protea SDK consistency:
  - remove hardcoded SDK URL from `usePTISdk`
  - use backend-provided widget/sdk config in PTI flows
- Optional SDK stub support for deterministic browser events:
  - `UserAssessmentCompleted`
  - `AddCreditCardCompleted`

## TDD Rules
- Keep test-first cycles for backend and frontend behavior changes.
- Refactor aggressively as config and integration paths are unified.

## Persistence and Queue Rules
- Preserve shared store interface usage and webhook jobs behavior from prior phases.
- E2E still runs on redis-backed store.

## Testing Requirements
- Local profile smoke tests for PTI enabled path.
- Browser e2e signup flow stability checks.
- Feature-file milestone for this phase (if appropriate):
  - no new mandatory API feature file today; keep all existing feature files green and add a dedicated local/browser feature file when introduced.
- Internal coverage gate: `go test -coverprofile=coverage.out ./internal/...` must report `>= 75.0%` total statements.

## Deliverables
- PTI can be switched between mock and real provider by env config only.
- No hardcoded external PTI SDK dependency in core local signup path.

## Acceptance Criteria
- Local environment uses mockpti without app code edits.
- Browser signup e2e can run reliably with mockpti-backed configuration.
- Prior phase tests remain green.
- Existing API feature files remain green end-to-end.
- Internal package total coverage is `>= 75.0%`.

## Verification
Run `make test` from the `go/mock/mockpti` directory.
Run `go test -coverprofile=coverage.out ./internal/... && go tool cover -func=coverage.out | grep total` and verify total coverage is `>= 75.0%`.
The phase is complete only when both checks pass.
