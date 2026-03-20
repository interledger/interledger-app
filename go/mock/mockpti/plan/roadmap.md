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
- Coverage gate: every phase must keep `go test -coverprofile=coverage.out ./internal/...` at `>= 75.0%` total statement coverage.
- Feature progression gate: each phase must make the next planned feature file pass end-to-end (when applicable).

## Phase Sequence
1. ~~Phase 1: Foundations and Signup Core~~ ✅ Done
2. ~~Phase 2: Wallets and Payment Information~~ ✅ Done
3. Phase 3: Transactions ← current
4. Phase 4: Webhook Jobs and Delivery
5. Phase 5: Local Integration and SDK Reliability

## Feature File Milestones
1. Phase 1 unlocks:
  - `features/service_health.feature`
  - `features/token_generation.feature`
  - `features/user_and_kyc.feature`
2. Phase 2 unlocks:
  - `features/wallet_and_payment_information.feature`
3. Phase 3 unlocks:
  - `features/transactions.feature`
4. Phase 4 unlocks:
  - `features/webhooks.feature`
5. Phase 5 has no additional mandatory feature file today; it stabilizes local/browser integration while keeping all prior feature files green.

## Documents
- `go/mock/mockpti/plan/phase-1-foundations-signup.md`
- `go/mock/mockpti/plan/phase-2-wallets-payment-info.md`
- `go/mock/mockpti/plan/phase-3-transactions.md`
- `go/mock/mockpti/plan/phase-4-webhook-jobs.md`
- `go/mock/mockpti/plan/phase-5-local-integration-sdk.md`

## Suggested PR Strategy
- One PR per phase.
- Do not start the next phase until the prior phase acceptance criteria pass.
- If a phase grows too large, split by endpoint groups but keep phase boundaries intact.

## Signup E2E Milestone
- Backend signup path becomes viable by end of Phase 4.
- Full deterministic browser signup e2e target is end of Phase 5.
