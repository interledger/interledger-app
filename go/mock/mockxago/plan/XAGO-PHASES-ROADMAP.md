# Xago Integration Phases: Complete Roadmap

## Executive Summary

This document provides a high-level overview of all Xago integration phases (5-11), their dependencies, timelines, and success criteria. Each phase has a detailed implementation plan in its respective file.

## Phase Overview

| Phase | Title | Prerequisites | Status |
|-------|-------|---------------|--------|
| **5** | MockXago Unit Testing & Feature Import | None (Foundation) | ✅ Complete |
| **6** | Local Docker Compose Setup | Phase 5 | ✅ Complete |
| **7** | Wallet E2E Integration for Xago Signup Processes | Phase 6 | ✅ Complete |
| **8** | Xago KYC Processes | Phase 7 | ✅ Complete |
| **9** | Xago Deposit Processes | Phase 8 | Planning |
| **10** | Xago P2P Payment Processes | Phase 9 | Planning |
| **11** | Xago Withdrawal Processes | Phase 10 | Planning |

## Phase Dependencies

```
   Phase 5 (MockXago Foundation)
   ↓
   Phase 6 (Local Docker Compose Setup) ← PREREQUISITE for all E2E tests!
   ↓
   Phase 7 (Signup)
   ↓
   Phase 8 (KYC)
   ↓
   Phase 9 (Deposits)
   ↓
   Phase 10 (P2P Payments) ← Critical correctness bellwether
   ↓
   Phase 11 (Withdrawals)
   ↓
   Full integration test & next phases
```

## Phase Details at a Glance

### Phase 5: MockXago Unit Testing & Feature Import (Foundation)
**File**: [phase-5-plan.md](phase-5-plan.md)

**What it does**:
- Imports 3 missing feature files from bosbaber/mockxago branch
- Implements Redis storage backend for MockXago
- Expands unit test coverage from 44% to 65%+ for `internal/` packages
- Sets up E2E BDD testing infrastructure with docker-compose and Redis

**Why it's critical**:
- All subsequent phases depend on MockXago being fully tested and operational
- Redis is essential for all integration tests
- 65%+ coverage ensures code quality baseline

**Must-pass validation**:
```bash
cd go/mock/mockxago
go test ./internal/... -coverprofile=coverage.out
go tool cover -func=coverage.out | grep total  # Must be >= 65%
go test ./testenv/... -v  # All 8 features must pass
```

**Exit criteria**: 8 feature files, 65%+ coverage, all tests passing

---

### Phase 6: Local Docker Compose Setup (Critical Infrastructure)
**File**: [phase-6-plan.md](phase-6-plan.md)

**What it does**:
- Adds `local/mockxago.yaml` Docker Compose service definition
- Configures backend service with all required XAGO_* environment variables (10 variables)
- Updates `local/wallet.yaml` with Xago provider configuration
- Sets up Redis database assignment (DB 4 for MockXago)
- Configures service networking (Traefik routing for mockxago.interledger.test)
- Documents health check and dependency ordering

**Why it's CRITICAL**:
- **Blocking issue** — Phase 7-11 cannot start without this
- E2E tests need fully configured local environment
- Backend needs Xago API endpoints to be accessible
- Redis DB isolation prevents conflicts with other services

**Must-pass validation**:
```bash
cd local
make down
make all  # Starts all services
# Wait ~30s for services to stabilize
sleep 30

# Verify services running
curl -sk https://mockxago.interledger.test/health
# Should return 200 OK

curl -sk https://backend.interledger.test/health | jq '.xago_status'
# Should show Xago provider configured

# Verify Redis connectivity
redis-cli -p 6379 -n 4 PING
# Should return PONG
```

**New/Modified files**:
- `local/mockxago.yaml` (new)
- `local/wallet.yaml` (modified - add XAGO_* vars)
- `local/docker-compose.yml` (may need minor updates for service ordering)

**Exit criteria**: 
- All services start successfully
- MockXago accessible at mockxago.interledger.test
- Backend configured with Xago endpoints
- Redis DB 4 available and isolated

---

### Phase 7: Wallet E2E Integration for Xago Signup Processes
**File**: [phase-7-plan.md](phase-7-plan.md)

**What it does**:
- Adds Xago-tagged signup scenario to `e2e/features/000-signup.feature`
- Adds mockxago service reference to feature file Background
- Updates `e2e/gatehub_signup.go` with Xago signup step registration
- Verifies frontend/backend/MockXago integration for Xago users

**Key insight**:
- Signup is provider-agnostic (shared code)
- Only differences: feature file Background line + step registration
- No new step files needed (signup flow works for both providers)

**Why it's important**:
- First E2E integration test validating full stack
- Proves frontend, backend, and MockXago work together
- Establishes pattern for subsequent Xago E2E tests
- Much simpler than initially thought (3-4 hours vs 6-10)

**Must-fail-to-proceed validation**:
```bash
cd e2e
go test -v -timeout 15m -args -tags="@xago&&@signup"
# All scenarios must pass, no undefined steps
```

**New/Modified files**:
- `e2e/features/000-signup.feature` (modified - add mockxago Background line + @xago scenario)
- `e2e/gatehub_signup.go` (modified - add step registration for Xago)

---

### Phase 8: Xago KYC Processes
**File**: [phase-8-plan.md](phase-8-plan.md)

**What it does**:
- Adds Xago KYC scenario to `e2e/features/001-kyc.feature`
- Modifies `e2e/gatehub_kyc.go` to add provider detection for South Africa users
- Routes South Africa users to MockXago mocked Persona iframe
- Verifies KYC completion and state transition to "accepted"

**Key insight**:
- KYC also uses provider-agnostic shared code pattern
- Detection: `if country == "South Africa" → route to Xago`
- No new step file needed (provider routing embedded in shared handler)
- Much simpler than initially thought (4-6 hours vs 8-12)

**Complexity**: Medium (Persona mock form submission)

**Why it's important**:
- KYC is prerequisite for deposits and withdrawals
- Persona mocking is proven in bosbaber/mockxago (reference implementation)
- Critical gateway workflow

**Must-fail-to-proceed validation**:
```bash
cd e2e
go test -v -timeout 20m -args -tags="@xago&&@kyc"
# All scenarios must pass, backend logs show KYC events
```

**New/Modified files**:
- `e2e/features/001-kyc.feature` (modified - add mockxago Background line + @xago scenario)
- `e2e/gatehub_kyc.go` (modified - add South Africa provider detection + iFillAndSubmitTheMockxagoiframe() helper)

---

### Phase 9: Xago Deposit Processes
**File**: [phase-9-plan.md](phase-9-plan.md)

**What it does**:
- Imports Xago deposit scenarios
- Implements deposit initiation and webhook handling
- Creates E2E step definitions for deposit flow
- Verifies balances update correctly after deposits
- Tests multi-currency deposit independence

**Complexity**: Low-Medium (straightforward balance updates)

**Why it's important**:
- Users need to fund wallets to do anything else
- Simpler than KYC, good confidence builder
- Foundation for P2P payments

**Must-fail-to-proceed validation**:
```bash
cd e2e
go test -v -timeout 20m -args -tags="@xago&&@deposit"
# All scenarios must pass, database shows updated balances
```

**New files**: `e2e/xago_deposit.go`

---

### Phase 10: Xago P2P Payment Processes
**File**: [phase-10-plan.md](phase-10-plan.md)

**What it does**:
- Imports Xago P2P payment scenarios
- Implements two-user payment flow (sender → recipient)
- Creates E2E step definitions for payment scenarios
- Verifies double-entry accounting (sender balance -, recipient balance +)
- Tests transaction history for both parties

**Complexity**: Medium-High (requires careful double-entry verification)

**Why it's CRITICAL**:
- **Bellwether phase** — if P2P works, financial accounting is correct
- If this fails, withdrawals (Phase 11) will also fail
- Proves twice-as-many balance updates are working (2 users affected per transaction)
- Tests recipient resolution and routing

**Must-fail-to-proceed validation**:
```bash
cd e2e
go test -v -timeout 30m -args -tags="@xago&&@p2p-payment"
# All scenarios must pass
# Verify: balance_before_A > balance_after_A
#         balance_before_B < balance_after_B
#         (decreased_A) == (increased_B)  ← Conservation of money!
```

**New files**: `e2e/xago_p2p_payment.go`

---

### Phase 11: Xago Withdrawal Processes
**File**: [phase-11-plan.md](phase-11-plan.md)

**What it does**:
- Imports Xago withdrawal scenarios
- Implements withdrawal initiation with destination specification
- Handles fee calculations and balance reconciliation
- Creates E2E step definitions for withdrawal flow
- Verifies balances update correctly (amount + fee deducted)
- Tests multiple withdrawal types (bank account, crypto address, etc.)

**Complexity**: Medium (fee calculations add complexity)

**Why it's important**:
- Final core workflow — users can now cash out
- Completes the financial cycle: deposit → transfer → withdraw
- Fee handling introduces new verification requirement

**Must-fail-to-proceed validation**:
```bash
cd e2e
go test -v -timeout 30m -args -tags="@xago&&@withdrawal"
# All scenarios must pass
# Verify: balance_before - amount - fee = balance_after
```

**Post-Phase validation** (Full Xago Journey):
```bash
cd e2e
go test -v -timeout 90m -args -tags="@xago"
# All phases: signup → kyc → deposit → p2p → withdraw
# Single user should complete full lifecycle
```

**New files**: `e2e/xago_withdrawal.go`

---

## Critical Success Factors

### For Each Phase
1. **All tests must pass** before proceeding to next phase
2. **Database state must be consistent** (no orphaned records)
3. **Backend logs must be clean** (no errors related to the phase)
4. **No undefined steps** in Godog test output

### Across All Phases
1. **Redis is mandatory** — Phase 5 must succeed
2. **Minimal backend changes** — Keep changes small for PR review
3. **Double-entry accounting** — Phase 10 is the bellwether
4. **Feature parity with bosbaber/mockxago** — Reference implementation guides all phases

## Testing Roadmap

### Phase 5 Validation
```bash
cd go/mock/mockxago
make test  # or: go test ./...
go tool cover -func=coverage.out | grep total  # >= 65%
```

### Phases 7-11 Validation (Common Pattern)
```bash
# For each phase N:
cd e2e
go test -v -timeout 20m -args -tags="@xago&&@TAG"  # where TAG=signup|kyc|deposit|p2p-payment|withdrawal
# Verify: ✓ All scenarios pass
#         ✓ Backend logs clean
#         ✓ Database state correct
#         ✓ Screenshots OK (no UI errors)
```

### Final Validation (After Phase 11)
```bash
cd e2e
go test -v -timeout 90m -args -tags="@xago"
# Should complete end-to-end Xago user journey without errors
```

## Implementation Strategy

### Recommended Approach
1. **Do Phases 5 sequentially** (foundation must be solid)
2. **Do Phase 6 as needed** (critical infrastructure prerequisite)
3. **Do Phases 7-11 in order** (each depends on previous)
4. **Run full integration test** after Phase 11 (before next phases)

### Branching Strategy
- One branch per phase: `stephan/phase-7`, `stephan/phase-8`, etc.
- Each phase is small enough (~2KB code changes) for 30-minute review
- Reference bosbaber/mockxago branch for patterns

## Common Patterns Across Phases

### Step Definition Pattern
```go
func (ctx *TestContext) iDoSomething(arg1 string) error {
    // Implementation
    return nil
}
```

### Validation Pattern
```bash
// Check: Scenario passes
go test -v -timeout TIME -args -tags="@TAG"

// Check: Backend processes events
docker logs backend | grep -i "keyword"

// Check: Database state
docker exec postgres psql -d backend -c "SELECT ... WHERE provider='xago';"
```

### Debugging Pattern
1. Check `e2e/debug/` for screenshots
2. Check backend/MockXago logs
3. Query database for state
4. Manually test API call (curl)

## Risk Mitigation Strategy

### Top Risks
1. **Persona mocking too complex (Phase 8)** → Reference bosbaber/mockxago implementation
2. **Double-entry accounting broken (Phase 10)** → Extensive verification assertions
3. **Fee calculations wrong (Phase 11)** → Test manually before E2E integration
4. **Backend routing doesn't support Xago** → Query codebase early in each phase

### Quality Gates
- Code coverage >= 65% (Phase 5)
- All tests pass (every phase)
- No hardcoded provider names (all phases, use environment/configuration)
- Database consistency validated (Phases 7-11)

## Integration with CI/CD

### Before PR Merge
- Run all tests for the phase: `go test ./... -v`
- Generate coverage report (if Phase 5)
- Run E2E scenarios: `go test ./e2e -v -args -tags="@TAG"`
- Verify database cleanup (no orphaned records)

### After Phase Completion
- Tag commit with phase number
- Document any deviations from plan
- Update dependencies if needed
- Note any patterns discovered for future phases

## Future Phases (Beyond Phase 11)

After Phase 11, potential future phases could include:
- **Phase 12**: Linked account management
- **Phase 13**: Advanced features (FX, batch operations, limits)
- **Phase 14**: Error handling and edge cases
- **Phase 15**: Performance and load testing

## Document References

- **Phase 5**: [phase-5-plan.md](phase-5-plan.md) — Foundation with unit tests
- **Phase 7**: [phase-7-plan.md](phase-7-plan.md) — Signup integration
- **Phase 8**: [phase-8-plan.md](phase-8-plan.md) — KYC workflows
- **Phase 9**: [phase-9-plan.md](phase-9-plan.md) — Deposits
- **Phase 10**: [phase-10-plan.md](phase-10-plan.md) — P2P Payments (critical!)
- **Phase 11**: [phase-11-plan.md](phase-11-plan.md) — Withdrawals

Each phase document contains:
- Detailed implementation tasks
- Testing strategies
- Success criteria
- Risk mitigation
- Estimated timeline

## Key Principles

1. **Tests must pass before proceeding** — Non-negotiable gate between phases
2. **Minimal backend changes** — Keep PRs small and reviewable
3. **Reference bosbaber/mockxago** — Proven implementation exists, learn from it
4. **Database state is source of truth** — Queries validate correctness
5. **Reuse patterns** — Step definitions, helpers, validation techniques
6. **Document deviations** — If implementation differs from plan, update docs

## Contact Points

When implementing phases:
1. **Phase 5 issues**: Review MockXago unit tests, check coverage gaps
2. **Phase 8 issues**: Reference bosbaber/mockxago Persona implementation
3. **Phase 10 issues**: Verify double-entry accounting math, check webhook ordering
4. **Phase 11 issues**: Understand fee structure from MockXago/Xago docs
5. **General questions**: Review corresponding phase plan document first

---

**Document Status**: Complete Phase Plans (5, 7-11)
**Date Created**: March 6, 2026
**Last Updated**: March 17, 2026
**Status**: Phases 5–8 complete; Phase 9 next
**Next Step**: Begin Phase 9 implementation per [phase-9-plan.md](phase-9-plan.md)
