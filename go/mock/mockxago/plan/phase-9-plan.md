# Phase 9 Implementation Plan: Xago Deposit Processes

## Overview

Phase 9 integrates Xago deposit workflows into the E2E test suite. This phase focuses on testing the deposit flow where users receive funds via Xago, with proper balance updates and transaction tracking.

**Phase Outcome**: E2E feature tests for Xago deposit processes pass, enabling users to fund their Xago wallets.

## Current Status

- ✓ Phase 5 complete: MockXago with 65%+ coverage and full feature tests
- ✓ Phase 7 complete: Xago signup scenarios integrated
- ✓ Phase 8 complete: Xago KYC workflows integrated
- ✓ GateHub deposit workflows exist as reference implementation

## Scope

### What's Included
- Xago deposit initiation from `bosbaber/mockxago` (e2e/features/002-deposit.feature)
- Deposit amount specification and currency selection
- Webhook-based balance updates
- Deposit confirmation and wallet balance validation
- Transaction history tracking

### What's NOT Included (Defer to Later Phases)
- P2P payments (Phase 10)
- Withdrawals (Phase 11)
- Advanced deposit features (e.g., multiple deposits, deposit fees)

## Key Differences: Xago Deposits vs GateHub Deposits

| Aspect | GateHub | Xago |
|--------|---------|------|
| **Deposit Model** | Vault funds to address | Subaccount receives transfer |
| **Webhook Event** | `core.deposit.completed` | `core.deposit.completed` (likely same) |
| **Balance Update** | Immediate or after webhook | After webhook callback |
| **Currency Support** | Multi-currency vaults | Multi-currency subaccounts |
| **Reference Routing** | XRPL memo field | Deposit reference/ID field |

## Implementation Tasks

### Task 1: Import Deposit Feature Scenarios (30-45 min)
**Objective**: Bring Xago deposit scenarios into feature files

**Action**:
1. Review `bosbaber/mockxago:e2e/features/002-deposit.feature` for Xago deposit scenarios
2. Add scenarios to `e2e/features/002-deposit.feature` with `@xago` tag
3. Ensure scenarios cover:
   - Deposit initiation (specifying amount and currency)
   - Deposit reference/reference ID
   - Balance before deposit
   - Webhook callback triggering balance update
   - Balance after deposit verification
   - Multiple deposits in different currencies
   - Deposit transaction visible in history

**Validation**:
- `grep "@xago" e2e/features/002-deposit.feature` shows deposit scenarios
- Feature file syntax valid: `go test ./e2e --dry-run -args -tags="@xago&&@deposit"`

### Task 2: Understand MockXago Deposit Mechanism (30-45 min)
**Objective**: Review how MockXago processes deposits

**Investigation actions**:
1. Review `bosbaber/mockxago:go/mock/mockxago/features/currencies_and_deposits.feature`
2. Examine `go/mock/mockxago/internal/handler/deposit_handler.go` (if exists) or transaction handler
3. Check MockXago storage for balance tracking
4. Document how deposits work in MockXago:
   - Deposit endpoint (if direct API)
   - Webhook delivery for deposit completion
   - Balance query endpoint

**Questions to answer**:
- What endpoint triggers a deposit? (e.g., `POST /deposits` or webhook-driven)
- What format does the deposit payload use?
- How does MockXago route funds to the correct subaccount?
- When are webhooks sent (immediate or delayed)?

**Output**: Document deposit flow understanding

### Task 3: Implement Deposit Step Definitions (1-1.5 hours)
**Objective**: Create E2E steps for Xago deposit scenarios

**Files to create**:
1. **New**: `e2e/xago_deposit.go` (separate provider-specific deposit steps file)
   - Step: "I request a deposit of {amount} {currency} to my wallet"
   - Step: "The deposit is received and processed"
   - Step: "My {currency} balance is {amount}"
   - Step: "The deposit appears in my transaction history"
   - Step: "I receive a deposit notification"
   - Step: "Multiple deposits are tracked independently per currency"

2. **Helpers**: Add to `e2e/` utilities
   - Function to initiate deposit via MockXago API
   - Function to query current balance
   - Function to poll for deposit webhook
   - Function to verify transaction in history

**Implementation approach**:
1. Deposit initiation: Call MockXago `POST /deposits` endpoint (or equivalent)
2. Wait for webhook: Poll backend webhook receipt OR poll balance endpoint
3. Verify balance: Query backend wallet balance API
4. Check history: Query transaction list endpoint

**Validation**:
- Step functions exist: `grep "^func" e2e/xago_deposit_steps.go | wc -l`
- No compile errors: `go build ./e2e/...`

### Task 4: Wire Deposits into Backend (30-60 min)
**Objective**: Ensure backend correctly processes Xago deposit webhooks

**Likely minimal changes**:
1. **Verify backend webhook endpoint** accepts Xago deposit events
   - Event: `core.deposit.completed` (or Xago variant)
   - Check if provider-agnostic or GateHub-hardcoded
   
2. **Check balance update logic**
   - Does webhook handler correctly update wallet balance?
   - Does it handle multi-currency correctly?
   - Does it log deposit transactions?

3. **No new backend endpoints needed** — reuse existing webhook handler

**Approach**:
- Query backend code for deposit webhook handler
- Search for provider routing logic
- If provider-agnostic: no changes needed
- If GateHub-specific: generalize if necessary

**Validation**:
- Backend starts without errors
- Webhook handler processes Xago events: `docker logs backend | grep "deposit"`
- Database shows balance updates for Xago accounts

### Task 5: Run Deposit Scenarios & Debug (1-1.5 hours)
**Objective**: Execute deposit scenarios and fix issues

**Actions**:
1. **Run scenarios**:
   ```bash
   cd e2e
   go test -v -timeout 20m -args -tags="@xago&&@deposit"
   ```

2. **Handle common issues**:
   - **Undefined steps**: Implement missing step definitions
   - **Deposit not received**: Check MockXago logs, verify webhook delivery
   - **Balance not updated**: Check backend webhook handler, verify database update
   - **Timing issues**: Add retry logic or longer waits if needed
   - **Multi-currency tracking**: Verify each currency balance independentally

3. **Debug techniques**:
   - Screenshots: `e2e/debug/`
   - MockXago logs: `docker logs mockxago | tail -50`
   - Backend logs: `docker logs backend | grep -i deposit`
   - Database: `psql -d backend -c "SELECT * FROM user_walances LIMIT 5;"`
   - Webhook receipt: Verify webhook actually reached backend

4. **Iterate** until all scenarios pass

**Validation**:
- All `@xago&&@deposit` scenarios pass
- No undefined steps or failures
- Balances correctly updated in database

## Testing Strategy

### Before Proceeding to Phase 10

All the following must pass:

```bash
cd e2e

# 1. All Xago deposit scenarios pass
go test -v -timeout 20m -args -tags="@xago&&@deposit"

# 2. Backend webhook processing successful
docker logs backend | grep -i "deposit.*completed"
# Should show at least one event per scenario

# 3. Database balances updated
docker exec postgres psql -d backend -c \
  "SELECT user_id, currency, balance FROM user_balances WHERE provider='xago';"
# Should show updated balances

# 4. No undefined steps
# (verify from test output - all steps should have implementations)
```

## Key Technical Considerations

### Deposit Flow Understanding
Before implementing, understand:
1. **How does MockXago trigger deposits?**
   - Is it a direct API call (E2E test calls it)?
   - Or does E2E test call backend, which calls MockXago?
   
2. **Webhook delivery**:
   - MockXago sends webhook to backend after processing deposit
   - Backend receives webhook and updates balance
   - E2E test polls for balance update confirmation

3. **Reference routing**:
   - Xago uses deposit "reference" to route funds to correct subaccount
   - E2E test must provide valid reference when initiating deposit

### Balance Precision
- Amounts typically use decimal precision (2 decimal places for fiat)
- XRP may use different precision
- Ensure test assertions handle floating-point correctly
- Database may store as string to avoid precision loss

### Webhook Timing
- Deposit webhook may be delayed (background job in MockXago)
- E2E test must retry polling for balance updates
- Timeout: 5-10 seconds for balance update after webhook
- Don't assume synchronous completion

## Implementation Notes

### Minimal Backend Changes Expected
- No new endpoints needed (reuse existing webhook handler)
- No UI changes (deposit form works for all providers)
- No database schema changes
- Only configuration changes if needed

### Reusable Patterns
- Balance query helper (can be reused for other balance checks)
- Webhook polling helper (can be reused for other async operations)
- Transaction history query (can be reused for other transaction checks)

### Error Handling
- Invalid deposit amount should error clearly
- Wrong currency should error or convert
- Invalid reference should return error from MockXago
- Network timeout should retry and fail gracefully

## Success Criteria

Phase 9 is **COMPLETE** when:

1. ✅ `e2e/features/002-deposit.feature` contains `@xago` tagged deposit scenarios
2. ✅ `e2e/*xago*deposit*.go` files exist with step implementations
3. ✅ All `@xago&&@deposit` scenarios pass: `go test ./e2e -v -args -tags="@xago&&@deposit"`
4. ✅ Backend logs show deposit webhooks received from MockXago
5. ✅ Database shows updated balances after deposit
6. ✅ No undefined step errors
7. ✅ Multi-currency deposits tracked independently

**All tests must pass before Phase 10 can proceed.**

## Risk Mitigation

| Risk | Mitigation |
|------|-----------|
| MockXago deposit mechanism unclear | Review bosbaber/mockxago feature files and handler code thoroughly |
| Webhook delivery timing issues | Implement polling with reasonable timeout (5-10s), not immediate checks |
| Backend doesn't recognize Xago deposits | Check webhook handler early, search for provider-specific logic |
| Balance precision issues | Store all amounts as strings in tests, compare with decimal tolerance |
| Reference routing fails | Document reference format, verify it's passed correctly to MockXago |

## Dependency on Previous Phases

- **Phase 5**: MockXago foundation working
- **Phase 7-8**: User can sign up and complete KYC (can now receive deposits)
- **Backend**: Webhook handler operational, balance update logic working

## Notes for Implementation Team

- **Simpler than KYC** — no Persona mocking, just balance updates
- **Reuse patterns** from Phase 7-8 (API calls, webhook polling, database queries)
- **Focus on balance verification** — this is the core test objective
- **Multi-currency is key** — ensure different currencies don't interfere with each other
- **Use deposits from multiple test scenarios** to build up history for later phase testing

---

**Phase**: 9 (Xago Deposits)
**Prerequisite**: Phase 8 (Xago KYC) — ✅ Complete
**Next Phase**: Phase 10 (Xago P2P Payments)
**Last Updated**: March 17, 2026
