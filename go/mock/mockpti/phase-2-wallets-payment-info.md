# Phase 2: Wallets and Payment Information

## Objective
Implement wallet and payment information APIs required by backend PTI onboarding flows.

## Scope
- Wallet endpoints:
  - `POST /users/{id}/wallets`
  - `GET /users/{id}/wallets`
  - `GET /users/{id}/wallets/{walletId}`
- Payment information endpoints:
  - `POST /users/{id}/payment-information`
  - `GET /users/{id}/payment-information/{id}`

## TDD Rules
- Add failing unit tests per behavior before implementation.
- Refactor aggressively after each green cycle.

## Persistence and Architecture Rules
- All reads/writes go through the shared store interface.
- Unit tests continue to run against memory store only.

## Testing Requirements
- Unit coverage for success and validation/error paths.
- E2E coverage for basic wallet/payment-information happy path using redis store.

## Deliverables
- Stable wallet/payment-information API behavior aligned with backend expectations.
- Updated feature tests and testenv support.

## Acceptance Criteria
- All phase endpoints pass unit and e2e scenarios.
- No leakage of storage concerns into handlers/domain logic.
- Existing Phase 1 tests remain green.
