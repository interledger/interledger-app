# Phase 11 Implementation Plan: Xago Withdrawal Processes

## Overview

Phase 11 integrates Xago withdrawal workflows into the E2E test suite. This is the final phase of Xago core functionality, enabling users to cash out their Xago wallet balances to external bank accounts or crypto addresses.

**Phase Outcome**: E2E feature tests for Xago withdrawal processes pass, with verified fund transfers from Xago wallets to destination accounts.

## IMPORTANT
The implementing agent is expected to compare the current git branch with the git `bosbaber/mockxago` branch to gain a better understanding of the task at hand. The current branch is probably more sophisticated and more up to date, but the tests are passing in the bosbaber/mockxago branch.

## Current Status

- ✓ Phase 5 complete: MockXago with 65%+ coverage and full feature tests
- ✓ Phase 7 complete: Xago signup scenarios integrated
- ✓ Phase 8 complete: Xago KYC workflows integrated
- ✓ Phase 9 complete: Xago deposit workflows integrated
- ✓ Phase 10 complete: Xago P2P payment workflows integrated
- ✓ GateHub withdrawal workflows exist as reference implementation

## Scope

### What's Included
- Xago withdrawal initiation from `bosbaber/mockxago` (e2e/features/004-withdrawal.feature)
- Withdrawal amount and destination specification
- Withdrawal fee calculation and deduction
- Balance verification before and after withdrawal
- Webhook-based withdrawal completion
- Withdrawal transaction tracking
- Error handling (insufficient balance, invalid destination)

### What's NOT Included (Defer to Later Phases)
- Advanced withdrawal features (e.g., scheduled withdrawals, batch withdrawals)
- KYC verification for withdrawal (assumed already done in Phase 8)
- Exchange rate/FX on withdrawals (if applicable)
- Linked account management (separate phase if applicable)

## Key Differences: Xago Withdrawals vs GateHub Withdrawals

| Aspect | GateHub | Xago |
|--------|---------|------|
| **Withdrawal Mechanism** | XRPL payment from address | Bank transfer / crypto send from subaccount |
| **Fee Model** | Optional withdrawal fee | Fees deducted by Xago or configured |
| **Destination Types** | XRPL address primarily | Bank account (varies by country) |
| **Balance Reconciliation** | Blockchain confirmation | Immediate deduction (webhook later) |
| **Waiting Period** | Blockchain timing | Bank processing time |
| **Webhook Events** | Transaction completion | Withdrawal completion / confirmation |

## Implementation Tasks

### Task 1: Import Withdrawal Feature Scenarios (30-45 min)
**Objective**: Bring Xago withdrawal scenarios into feature files

**Action**:
1. Review `bosbaber/mockxago:e2e/features/004-withdrawal.feature` for Xago withdrawal scenarios
2. Add scenarios to `e2e/features/004-withdrawal.feature` with `@xago` tag
3. Ensure scenarios cover:
   - Withdrawal initiation (amount, currency, destination)
   - Destination type specification (bank account, crypto address, etc.)
   - Balance before withdrawal
   - Balance after withdrawal (including fees)
   - Fee calculation verification
   - Withdrawal status tracking (pending → completed)
   - Withdrawal appears in transaction history
   - Error cases:
     - Insufficient balance
     - Invalid destination
     - Withdrawal exceeds limits
     - Network/backend errors

**Validation**:
- `grep "@xago" e2e/features/004-withdrawal.feature` shows withdrawal scenarios
- Feature file syntax valid: `go test ./e2e --dry-run -args -tags="@xago&&@withdrawal"`

### Task 2: Understand Xago Withdrawal Mechanism (45-60 min)
**Objective**: Review how withdrawals work in bosbaber/mockxago

**Investigation actions**:
1. **MockXago withdrawal flow** (`bosbaber/mockxago`):
   - Review `go/mock/mockxago/features/deposits_and_webhooks.feature` for withdrawal scenarios
   - Check `internal/handler/` for withdrawal endpoints
   - Understand fee calculations in MockXago

2. **Backend withdrawal flow** (`bosbaber/mockxago:backend`):
   - How does backend request withdrawals from Xago?
   - How are fees applied and stored?
   - How are balance updates handled?

3. **Destination handling**:
   - What destination types does Xago support?
   - How does E2E test specify destination?
   - Validation: what makes a destination invalid?

**Key questions to answer**:
- What endpoint triggers a withdrawal? (e.g., `POST /withdrawals`)
- How are fees calculated? (percentage, fixed amount, or both?)
- When is balance deducted? (immediately or after webhook?)
- What webhook events signal successful withdrawal?
- What are the withdrawal status states?

**Documentation to produce**:
- Withdrawal flow: User requests amount → fee calculated → balance deducted → webhook sent → completion
- Fee calculation example: 100 ZAR with 2% fee = 98 ZAR received (2 ZAR fee)
- Status states: pending | processing | completed | failed | reversed

### Task 3: Implement Withdrawal Step Definitions (1.5-2 hours)
**Objective**: Create E2E steps for Xago withdrawal scenarios

**Files to create**:
1. **New**: `e2e/xago_withdrawal.go` (provider-specific withdrawal steps)
   - Step: "I initiate a withdrawal of {amount} {currency} to {destination}"
   - Step: "The withdrawal fee is {fee}"
   - Step: "My balance decreases by {amount} plus {fee}"
   - Step: "The withdrawal is processed"
   - Step: "The withdrawal appears in my transaction history"
   - Step: "The withdrawal destination is {destination}"
   - Step: "I cannot withdraw more than my available balance"
   - Step: "An invalid destination is rejected"

2. **Helpers**: Add to `e2e/` utilities
   - Function to initiate withdrawal (call backend withdrawal API)
   - Function to specify withdrawal destination
   - Function to get withdrawal status
   - Function to poll for withdrawal completion webhook
   - Function to verify fee deduction

**Implementation approach**:
1. Get initial balance
2. Initiate withdrawal with destination
3. Verify balance decreased (amount + fee)
4. Poll for withdrawal completion webhook
5. Verify transaction appears in history

**Validation**:
- Step functions exist: `grep "^func" e2e/xago_withdrawal_steps.go | wc -l`
- No compile errors: `go build ./e2e/...`

### Task 4: Wire Withdrawals into Backend (30-45 min)
**Objective**: Ensure backend correctly handles Xago withdrawal workflows

**Investigation actions**:
1. **Find withdrawal endpoint**:
   - Check if `POST /withdrawals` exists and works for Xago
   - Verify it accepts destination parameter

2. **Check fee configuration**:
   - Does backend apply Xago withdrawal fees?
   - Are fees configurable per currency?
   - Are fees passed to MockXago or calculated in backend?

3. **Check webhook handler**:
   - Does backend process Xago withdrawal completion webhooks?
   - Does it mark withdrawal as completed?
   - Does it track withdrawal in transaction history?

4. **Check balance update**:
   - Is balance deducted when withdrawal is initiated?
   - Or when webhook is received?
   - What happens if withdrawal fails after balance deduction?

**Approach**:
- Query backend withdrawal handler code
- Test manually: withdrawal request → MockXago → webhook → transaction recorded
- Verify fee logic matches expected behavior

**Likely minimal changes required** (this should be mostly implemented if phases 7-10 work)

**Validation**:
- Backend starts without errors
- Withdrawal endpoint exists and responds
- Fee calculations are correct
- Webhooks trigger balance/transaction updates

### Task 5: Handle Fee Calculations (30-45 min)
**Objective**: Ensure withdrawal fees are correctly calculated and applied

**Key questions**:
1. **Fee structure**:
   - Fixed amount per withdrawal? (e.g., 50 ZAR)
   - Percentage-based? (e.g., 2% of amount)
   - Both (e.g., max(50 ZAR, 2% of amount))?
   - Different per currency?

2. **Fee ownership**:
   - Is fee deducted from sender's balance?
   - Does fee go to Xago, or backend merchant account?
   - How is fee tracked in transactions?

3. **Fee display**:
   - Should E2E test know fee in advance?
   - Or discover it after withdrawal is created?
   - How do we verify fee is correct without hardcoding?

**Action**:
1. Review bosbaber/mockxago fee calculation logic
2. Verify E2E has correct fee expectations
3. Add fee verification step to scenarios

**Validation**:
- Fees correctly deducted: `balance_before - amount - fee = balance_after`
- Fee calculation matches logic: test multiple amounts and currencies

### Task 6: Test Withdrawal Scenarios & Debug (1.5-2 hours)
**Objective**: Execute withdrawal scenarios and fix issues

**Actions**:
1. **Run scenarios**:
   ```bash
   cd e2e
   go test -v -timeout 30m -args -tags="@xago&&@withdrawal"
   ```

2. **Handle common issues**:
   - **Undefined steps**: Implement missing step definitions
   - **Invalid destination**: Verify destination validation
   - **Insufficient balance**: Verify error handling for low balance
   - **Fee calculation wrong**: Check logic in backend and MockXago
   - **Balance not updated**: Check webhook delivery
   - **Withdrawal not completed**: Check status polling logic

3. **Debug techniques**:
   - Screenshots: `e2e/debug/`
   - Backend logs: `docker logs backend | grep -i "withdrawal"`
   - MockXago logs: `docker logs mockxago | grep -i "withdrawal"`
   - Database queries:
     - Transaction status: `SELECT status FROM transactions WHERE type='withdrawal'`
     - Fee deductions: `SELECT balance FROM user_balances BEFORE/AFTER`
   - Webhook receipt: Verify webhooks reached backend

4. **Iterate** until all scenarios pass

**Validation**:
- All `@xago&&@withdrawal` scenarios pass
- No undefined steps
- Balances correctly updated (with fee deductions)
- Transaction history contains withdrawal records

## Testing Strategy

### Before Proceeding to Post-Phase 11

All the following must pass:

```bash
cd e2e

# 1. All Xago withdrawal scenarios pass
go test -v -timeout 30m -args -tags="@xago&&@withdrawal"

# 2. Backend withdrawal processing successful
docker logs backend | grep -i "withdrawal\|cashout"
# Should show withdrawal completion events

# 3. Balances updated correctly (including fee deduction)
docker exec postgres psql -d backend -c \
  "SELECT user_id, currency, balance FROM user_balances \
   WHERE provider='xago' ORDER BY user_id;"
# Should show decreased balance (amount + fee)

# 4. Withdrawal transactions recorded
docker exec postgres psql -d backend -c \
  "SELECT user_id, amount, fee, status FROM transactions \
   WHERE provider='xago' AND type='withdrawal';"
# Should show completed withdrawals with fees

# 5. No undefined steps
# (verify from test output)

# 6. Full cycle test (optional)
# Run all @xago scenarios in order: signup → kyc → deposit → p2p → withdrawal
go test -v -timeout 60m -args -tags="@xago"
```

## Key Technical Considerations

### Fee Precision
- Fees may be small fractions (e.g., 0.5%)
- Database may store as decimal or string to avoid floating-point errors
- Test assertions should use tolerance for floating-point comparisons

### Destination Validation
- Destination format depends on type (IBAN for bank, address for crypto)
- Invalid destinations should be rejected early
- E2E test should handle both valid and invalid cases

### Balance Ordering Assumption
- Balance should be checked BEFORE fee deduction in most systems
- When checking "insufficient balance", compare with: `balance < (amount + fee)`
- Test should verify this ordering

### Webhook Timing
- Withdrawal completion webhook may be delayed
- E2E test should poll for completion (timeout 5-10s)
- Don't assume synchronous completion

## Implementation Notes

### Minimal Backend Changes Expected
- No new database schema changes
- No new API endpoints (withdrawal endpoint should exist)
- Reuse existing transaction tracking
- May need fee configuration updates

### Why This is the Final Phase
- After withdrawals, all core Xago workflows are complete
- Future phases would be:
  - Advanced features (linked accounts, batch operations)
  - FX/multi-currency complications
  - Edge cases and error scenarios
  - Performance and load testing

### Reusable Patterns
- Destination validation helper (can be reused for wire transfers)
- Fee calculation verification (can be reused for other fee-bearing operations)

## Success Criteria

Phase 11 is **COMPLETE** when:

1. ✅ `e2e/features/004-withdrawal.feature` contains `@xago` tagged withdrawal scenarios
2. ✅ `e2e/*xago*withdrawal*.go` files exist with step implementations
3. ✅ All `@xago&&@withdrawal` scenarios pass: `go test ./e2e -v -args -tags="@xago&&@withdrawal"`
4. ✅ Backend logs show withdrawal processing and completion
5. ✅ Database shows:
   - Balances decreased by (amount + fee)
   - Transactions recorded for withdrawals
   - Fee amounts tracked correctly
6. ✅ No undefined step errors
7. ✅ Withdrawal destinations accepted/rejected correctly

**All tests must pass before calling Xago E2E integration complete.**

## Post-Phase 11: Full Xago Integration Test

Once Phase 11 is complete, run the full Xago lifecycle test:

```bash
cd e2e
go test -v -timeout 90m -args -tags="@xago"
```

This should:
1. Create new Xago user (Phase 7)
2. Complete KYC (Phase 8)
3. Receive deposit (Phase 9)
4. Send P2P payment (Phase 10)
5. Withdraw funds (Phase 11)

All steps should complete without errors, and the user should have a complete transaction history.

## Risk Mitigation

| Risk | Mitigation |
|------|-----------|
| Fee calculation logic complex | Document fee formula from MockXago, verify with examples |
| Destination validation missing | Check backend validation early, add to MockXago if needed |
| Balance precision issues | Use string/decimal types, avoid floating-point arithmetic |
| Withdrawal status unclear | Document state machine: pending → processing → completed |
| Backend fee handling different from expected | Test manually with curl before E2E integration |

## Dependency on Previous Phases

- **Phase 5**: MockXago foundation working
- **Phase 7-10**: Users can sign up, complete KYC, receive deposits, send P2P payments
- **Backend**: Withdrawal endpoint functional, fee calculation implemented, webhook handler operational

## Notes for Implementation Team

- **This is the final core phase** — after this, Xago is ready for more advanced work
- **Full lifecycle test is important** — run it to verify all pieces work together
- **Fee calculations are tricky** — test thoroughly with multiple amounts and currencies
- **Destination handling varies by country** — verify your test destinations are valid
- **Review bosbaber/mockxago thoroughly** — it works, learn from proven implementation
- **Document any differences from GateHub** — this helps future maintainers understand the system

---

**Phase**: 11 (Xago Withdrawals)
**Prerequisite**: Phase 10 (Xago P2P Payments)
**Next Step**: Run full Xago integration test, then plan advanced features
**Last Updated**: March 6, 2026

## Questions for Clarification

Before starting Task 2, determine:
1. What fee structure does Xago use? (fixed, percentage, tiered?)
2. What destination types are supported? (bank account, crypto address, both?)
3. Are fees deducted immediately or after webhook?
4. What withdrawal status states and transitions exist?
5. How does fee notification work (shown in advance, discovered after request, both)?
