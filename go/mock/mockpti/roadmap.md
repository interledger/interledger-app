# Mock PTI Implementation Roadmap

## Purpose
This roadmap breaks the mock PTI implementation into five reviewable phases with clear PR boundaries.

## Global Constraints (apply to all phases)
- Use strict TDD for every behavior: write failing unit tests first, implement minimal code, then refactor aggressively.
- Keep all persistence behind one store interface with two implementations:
  - memory store (unit tests)
  - redis store (e2e tests)
- Do not bypass the store interface from domain logic.
- Outgoing webhook HTTP calls must use a persisted jobs mechanism (no direct fire-and-forget HTTP from request handlers/domain code).

## Phase Sequence
1. Phase 1: Foundations and Signup Core
2. Phase 2: Wallets and Payment Information
3. Phase 3: Transactions
4. Phase 4: Webhook Jobs and Delivery
5. Phase 5: Local Integration and SDK Reliability

## Documents
- `go/mock/mockpti/phase-1-foundations-signup.md`
- `go/mock/mockpti/phase-2-wallets-payment-info.md`
- `go/mock/mockpti/phase-3-transactions.md`
- `go/mock/mockpti/phase-4-webhook-jobs.md`
- `go/mock/mockpti/phase-5-local-integration-sdk.md`

## Suggested PR Strategy
- One PR per phase.
- Do not start the next phase until the prior phase acceptance criteria pass.
- If a phase grows too large, split by endpoint groups but keep phase boundaries intact.

## Signup E2E Milestone
- Backend signup path becomes viable by end of Phase 4.
- Full deterministic browser signup e2e target is end of Phase 5.
