# Phase 10 Implementation Plan: Xago P2P Payment Processes

## Overview

Phase 10 integrates Xago peer-to-peer (P2P) payment workflows into the E2E test suite. This is the most complex phase so far, involving transfers between two Xago users with proper balance reconciliation and transaction tracking.

**Phase Outcome**: E2E feature tests for Xago P2P payments pass, with verified fund transfers between users maintaining accounting integrity.

## Current Status

- ✓ Phase 5 complete: MockXago with 65%+ coverage and full feature tests
- ✓ Phase 7 complete: Xago signup scenarios integrated
- ✓ Phase 8 complete: Xago KYC workflows integrated
- ✓ Phase 9 complete: Xago deposit workflows integrated
- ✓ GateHub P2P payment workflows exist as reference implementation

## Scope

### What's Included
- Xago P2P payment initiation from `bosbaber/mockxago` (e2e/features/003-p2p-payment.feature)
- Payment amount, currency, and recipient selection
- Sender and recipient balance verification before and after payment
- Webhook-based balance updates for both parties
- Transaction history tracking for both sender and recipient
- Payment status transitions (pending → completed)

### What's NOT Included (Defer to Later Phases)
- Withdrawals (Phase 11)
- Exchange rates / FX conversions (may be in later phases)
- Payment failures or rejections (Phase 10 is happy path)

## Key Differences: Xago P2P vs GateHub P2P

| Aspect | GateHub | Xago |
|--------|---------|------|
| **Transfer Mechanism** | XRPL payment to address | Xago subaccount transfer (internal) |
| **Balance Ledger** | Double-entry (Pacioli) | Xago internal balance tracking |
| **Webhook Events** | Transaction events | Transfer completion events |
| **ILP Integration** | May involve ILP routing | Direct Xago transfer (no ILP needed for same currency) |
| **Recipient Resolution** | Email or XRPL address | Xago subaccount ID or email |
| **Concurrency** | Blockchain timing | Instant Xago processing |

## Implementation Tasks

### Task 1: Import P2P Payment Feature Scenarios (30-45 min)
**Objective**: Bring Xago P2P payment scenarios into feature files

**Action**:
1. Review `bosbaber/mockxago:e2e/features/003-p2p-payment.feature` for Xago payment scenarios
2. Add scenarios to `e2e/features/003-p2p-payment.feature` with `@xago` tag
3. Ensure scenarios cover:
   - Two Xago users with adequate balances
   - Payment initiation (amount, currency, recipient)
   - Sender balance decreases
   - Recipient balance increases
   - Payment appears in both users' transaction histories
   - Webhook callbacks for both parties (if applicable)
   - Error cases (insufficient balance, invalid recipient)

**Validation**:
- `grep "@xago" e2e/features/003-p2p-payment.feature` shows payment scenarios
- Feature file syntax valid: `go test ./e2e --dry-run -args -tags="@xago&&@p2p-payment"`

### Task 2: Review bosbaber/mockxago P2P Implementation (1-1.5 hours)
**Objective**: Thoroughly understand how P2P transfers work in bosbaber/mockxago

**Key things to understand**:
1. **bosbaber/mockxago:go/mock/mockxago/features/transactions.feature**
   - How are P2P transfers tested at MockXago level?
   - What endpoints are involved?

2. **bosbaber/mockxago:e2e/features/003-p2p-payment.feature at source**
   - Detailed scenario structure
   - Step definitions referenced
   - What backend endpoints are called?

3. **bosbaber/mockxago:go/mock/mockxago/internal/handler/**
   - How does MockXago handle transfer requests?
   - How are recipient subaccounts resolved?
   - How are balance updates triggered?

4. **bosbaber/mockxago:bosbaber/mockxago:backend/**
   - How does backend handle Xago payment webhooks?
   - How are balances updated for sender and recipient?

**Documentation to produce**:
- Flow diagram: User A initiates payment → MockXago processes → Webhooks sent → Balances updated for A and B
- Endpoint list: Which backend/MockXago endpoints are called
- State transitions: Payment states (pending → completed)

### Task 3: Implement P2P Payment Step Definitions (1.5-2 hours)
**Objective**: Create E2E steps for Xago P2P payment scenarios

**Files to create**:
1. **New**: `e2e/xago_p2p_payment.go` (provider-specific P2P payment steps)
   - Step: "User {userA} sends {amount} {currency} to User {userB}"
   - Step: "{userA}'s balance decreases by {amount}"
   - Step: "{userB}'s balance increases by {amount}"
   - Step: "The payment appears in {userA}'s transaction history as sent"
   - Step: "The payment appears in {userB}'s transaction history as received"
   - Step: "User {userA} cannot send more than their available balance"
   - Step: "The payment is confirmed and completed"

2. **Helpers**: Add to `e2e/` utilities
   - Function to initiate P2P payment (call backend payment API)
   - Function to get user's current balance
   - Function to poll for balance updates after payment
   - Function to query transaction history
   - Function to resolve user to wallet ID / subaccount ID for recipient routing

**Implementation approach**:
1. Get initial balances for both users
2. Initiate payment via backend API (not MockXago directly)
3. Poll for balance changes (with timeout 5-10s)
4. Verify recipient received funds
5. Verify transaction appears in both histories

**Validation**:
- Step functions exist: `grep "^func" e2e/xago_p2p_payment_steps.go | wc -l`
- No compile errors: `go build ./e2e/...`

### Task 4: Wire P2P Payments into Backend (30-60 min)
**Objective**: Ensure backend correctly handles Xago P2P payment workflows

**Investigation actions**:
1. **Find payment endpoint**: Where does backend accept payment requests?
   - Likely: `POST /payments` or `POST /transfers`
   - Verify it works for Xago users

2. **Check provider routing**: Does backend know how to route Xago payments?
   - Must identify recipient's Xago subaccount
   - Must call MockXago transfer endpoint
   - Must handle response and webhooks

3. **Check webhook handler**: Does backend process Xago transfer completion webhooks?
   - Both sender and recipient should get webhook
   - Balances should update for both immediately

4. **Verify transaction tracking**: Are transactions created for both parties?

**Approach**:
- Query backend payment handler code
- Search for provider-specific routing
- Test manually: backend payment request → MockXago → webhook → balance update

**Likely minimal changes required** (this should be mostly implemented if phases 7-9 work)

**Validation**:
- Backend starts without errors
- Payment initiation succeeds: curl or API call
- Webhooks received: `docker logs backend | grep -i "transfer\|payment"`
- Both users' balances updated: database query

### Task 5: Handle Recipient Resolution (30-45 min)
**Objective**: Ensure E2E can correctly identify and route payments to recipient

**Key questions**:
1. How does backend identify the recipient?
   - By email?
   - By wallet ID?
   - By Xago subaccount ID?

2. How does MockXago routing work?
   - Must accept transfer to specific subaccount
   - Must update both sender and recipient balances

**Action**:
1. Review backend payment endpoint to understand recipient parameter
2. Ensure E2E test provides correct recipient identifier in payment request
3. Handle both cases:
   - Same currency transfer (simpler)
   - Different currency transfer (if applicable for Phase 10)

**Validation**:
- Test payment to known recipient (from same test session)
- Verify recipient identified correctly
- Verify transfer routed to correct subaccount

### Task 6: Run P2P Payment Scenarios & Debug (1.5-2 hours)
**Objective**: Execute payment scenarios and fix issues

**Actions**:
1. **Run scenarios**:
   ```bash
   cd e2e
   go test -v -timeout 30m -args -tags="@xago&&@p2p-payment"
   ```

2. **Handle common issues**:
   - **Undefined steps**: Implement missing step definitions
   - **Recipient not found**: Verify recipient resolution logic
   - **Balance not updated**: Check webhook delivery and handler
   - **Insufficient balance error**: Verify error handling
   - **Timing issues**: Add retry/polling logic
   - **Both balances not updated**: Check both sender and recipient webhook handlers

3. **Debug techniques**:
   - Screenshots: `e2e/debug/`
   - Backend logs: `docker logs backend | grep -i "payment\|transfer"`
   - MockXago logs: `docker logs mockxago | tail -50`
   - Database queries: Both users' balances before/after
   - Webhook receipt: Verify webhooks reached backend for both parties

4. **Iterate** until all scenarios pass

**Validation**:
- All `@xago&&@p2p-payment` scenarios pass
- No undefined steps
- Both users' balances correctly updated
- Transaction history correct for both parties

## Testing Strategy

### Before Proceeding to Phase 11

All the following must pass:

```bash
cd e2e

# 1. All Xago P2P payment scenarios pass
go test -v -timeout 30m -args -tags="@xago&&@p2p-payment"

# 2. Backend payment processing successful
docker logs backend | grep -i "payment.*completed\|transfer.*completed"
# Should show events for both sender and recipient

# 3. Both users' balances updated correctly
docker exec postgres psql -d backend -c \
  "SELECT user_id, currency, balance FROM user_balances \
   WHERE provider='xago' ORDER BY user_id;"
# Should show decreased balance for sender, increased for recipient

# 4. Transaction history for both users
docker exec postgres psql -d backend -c \
  "SELECT user_id, transaction_type, amount FROM transactions \
   WHERE provider='xago' AND status='completed';"
# Should show corresponding transactions for both parties

# 5. No undefined steps
# (verify from test output)
```

## Key Technical Considerations

### Double-Entry Accounting
- This is CRITICAL for P2P payments
- When User A sends to User B:
  - User A's balance: -amount
  - User B's balance: +amount
  - Total balance: unchanged (conservation of money)
- Verify this invariant in all tests

### Recipient Resolution
- Backend must reliably identify the recipient
- Email-based: lookup user by email in database
- Wallet ID: direct lookup
- Subaccount ID: verify belongs to intended recipient
- Clear error if recipient not found or ambiguous

### Webhook Ordering
- Sender and recipient may receive webhooks in any order
- E2E test must not assume specific ordering
- Must verify BOTH webhooks received before checking final balances

### Concurrency & Race Conditions
- If scenarios run in parallel, ensure test isolation
- Each test user should have unique balances
- Database queries should filter by user/wallet

## Implementation Notes

### Minimal Backend Changes Expected
- No new endpoints needed (reuse existing payment endpoint)
- No UI changes (payment form works for all providers)
- May need minor routing logic refinement (route payments to MockXago correctly)
- Webhook handler should already be provider-agnostic

### Why This Phase is Critical
- P2P is proof that MockXago can maintain double-entry accounting
- If double-entry breaks here, withdrawals (Phase 11) will also fail
- This is the bellwether for Xago financial correctness

### Reusable Patterns
- Recipient resolution helper (can be reused for withdrawals)
- Multi-user balance verification (can be reused for group scenarios)

## Success Criteria

Phase 10 is **COMPLETE** when:

1. ✅ `e2e/features/003-p2p-payment.feature` contains `@xago` tagged P2P scenarios
2. ✅ `e2e/*xago*p2p*.go` files exist with step implementations
3. ✅ All `@xago&&@p2p-payment` scenarios pass: `go test ./e2e -v -args -tags="@xago&&@p2p-payment"`
4. ✅ Backend logs show payment completion for both sender and recipient
5. ✅ Database shows:
   - Sender balance decreased
   - Recipient balance increased
   - Transactions recorded for both parties
6. ✅ No undefined step errors
7. ✅ Double-entry accounting verified (total balance unchanged)

**All tests must pass before Phase 11 can proceed.**

## Risk Mitigation

| Risk | Mitigation |
|------|-----------|
| Recipient resolution logic complex | Query backend code early, test manually before E2E integration |
| Webhook ordering issues | Implement polling for both webhooks, not sequential assumptions |
| Double-entry accounting broken | Add assertions to verify math: before + after = balanced |
| Concurrency/race conditions | Use distinct test users per scenario, verify database isolation |
| Backend payment endpoint doesn't support Xago | Check provider routing early, test with mock calls first |

## Dependency on Previous Phases

- **Phase 5**: MockXago foundation working
- **Phase 7-9**: Users can sign up, complete KYC, receive deposits (necessary preconditions)
- **Backend**: Payment endpoint functional, webhook handler operational, double-entry accounting working

## Notes for Implementation Team

- **This is THE critical test** — if P2P works, Xago financial system is sound
- **Double-entry accounting is non-negotiable** — verify invariants constantly
- **Review bosbaber/mockxago thoroughly** — it works, learn from it
- **Use multiple test scenarios** to build confidence
- **Database queries are debugging gold** — use them liberally
- **This is your bellwether phase** — success here means withdrawals should work cleanly in Phase 11

---

**Phase**: 10 (Xago P2P Payments)
**Prerequisite**: Phase 9 (Xago Deposits)
**Next Phase**: Phase 11 (Xago Withdrawals)
**Last Updated**: March 6, 2026

## Questions for Clarification

Before starting Task 2, determine:
1. Does backend payment endpoint already exist and work for all providers?
2. How does MockXago routing work for P2P transfers?
3. Are both sender and recipient webhooks fired, or just one?
4. Is double-entry accounting already implemented in backend, or needs to be verified?
