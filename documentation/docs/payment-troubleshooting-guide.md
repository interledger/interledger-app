# Payment Troubleshooting Guide

**Audience:** Support staff, operations, first-line debugging  
**Time to read:** 15 minutes

> **Debugging framework.** Systematic approach to resolving payment issues.

**Related documents:**

- [Payments Guide](payments-explainer.md) — Overview and navigation hub
- [Ledger System Architecture](ledger-system-explainer.md) — Understanding balance discrepancies
- [Transaction Types Reference](transaction-types-explainer.md) — Transaction field details
- [Provider Payments Guide](provider-payments-guide.md) — Provider-specific debugging paths
- [KYC Explainer](kyc-explainer.md) — KYC gating that can block payment flows
- [Concepts](concepts.md) — Core terminology

**Quick Navigation:**

- **Payment stuck?** → See [Problem 1: Payment Stuck in Processing](#problem-1-payment-appears-stuck-in-processing)
- **Balance wrong?** → See [Problem 2: Balance Mismatch](#problem-2-balance-is-wrong-shows-100-but-should-be-150)
- **Recipient didn't receive?** → See [Problem 3: Missing Receipt](#problem-3-sender-sent-money-but-recipient-didnt-receive-it)
- **Ledger discrepancy?** → See [Problem 4: Ledger Mismatch](#problem-4-our-balance-doesnt-match-providers-balance)
- **Using Temporal?** → See [Temporal Debugging Guide](#temporal-debugging-guide)

---

## Introduction

When something goes wrong with a payment, follow this order:
1) verify provider status,
2) compare with our records,
3) fix the mismatch at the source.

Don't guess — follow the evidence trail.

```mermaid
graph TD
    Problem["User reports:<br/>Payment stuck?<br/>Balance wrong?<br/>Money missing?"]
    
    Problem -->|Step 1| CheckData{"Do our records<br/>match provider's<br/>records?"}
    
    CheckData -->|No| FindDiff["Find the<br/>discrepancy"]
    CheckData -->|Yes| CheckTx["Check transaction<br/>status with provider"]
    
    FindDiff --> WhichOne{"Who's right?<br/>Us or provider?"}
    WhichOne -->|Provider<br/>authoritative| Update["Update our<br/>records to match<br/>provider"]
    WhichOne -->|Us<br/>correct| Retry["Retry or<br/>escalate to<br/>provider"]
    WhichOne -->|Unknown| Escalate["Investigate logs<br/>Check webhooks<br/>Manual review"]
    
    CheckTx --> TxStatus{"What status<br/>does provider<br/>report?"}
    TxStatus -->|Completed| MarkComplete["Mark transaction<br/>complete in our<br/>system"]
    TxStatus -->|Pending| CheckWebhook["Did we receive<br/>a webhook?"]
    TxStatus -->|Failed| MarkFailed["Mark failed<br/>Investigate why"]
    
    CheckWebhook -->|Yes| Strange["Strange mismatch:<br/>Webhook says<br/>complete but<br/>provider says<br/>pending"]
    CheckWebhook -->|No| Wait["Wait 20 min for<br/>provider to process<br/>then check again"]
    
    Strange --> Escalate
    Wait --> CheckAgain["Poll provider<br/>again"]
    CheckAgain --> TxStatus
    
    style Problem fill:#ff6b6b,stroke:#cc0000,color:#fff
    style FindDiff fill:#ffd93d,stroke:#cc8800,color:#000
    style Update fill:#70d070,stroke:#008800,color:#fff
    style Retry fill:#70d070,stroke:#008800,color:#fff
    style MarkComplete fill:#70d070,stroke:#008800,color:#fff
    style MarkFailed fill:#ff9999,stroke:#cc0000,color:#fff
    style Escalate fill:#ffd93d,stroke:#cc8800,color:#000
```

---

## Common Problems & Solutions

### Problem 1: "Payment appears stuck in 'Processing'"

**Symptoms:**
- User initiated payment hours ago
- Status still shows "Sending..."
- Balance hasn't changed

### Investigation Checklist

**Step 1: Check if transaction exists**

```bash
# Query our database
SELECT * FROM transactions 
WHERE id = 'txn_...' OR user_id = 'user_...'
ORDER BY created_at DESC
LIMIT 10;
```

- Transaction exists? ✓ Continue
- No transaction? ✗ Payment never created (UI bug or network issue)

**Step 2: Check transaction status**

```bash
# What status do we have?
Transaction 1:
├─ Status: "pending" (expected for stuck)
├─ Provider ID: "txn_abc123" (GateHub ID)
├─ Created At: 2 hours ago
└─ Provider Fee: £2.00
```

**Step 3: Check with provider (most important)**

```bash
# Query GateHub (or other provider) directly
GateHub API: GET /core/v1/transactions/{txn_abc123}

Response:
{
  "id": "txn_abc123",
  "status": 1,          ← Still pending!
  "amount": "100.00",
  "fee": "2.00"
}

OR

Response:
{
  "id": "txn_abc123",
  "status": 100,        ← Actually completed!
  "amount": "100.00",
  "fee": "2.00"
}
```

### Root Causes & Fixes

**If Provider Says: Pending**

```
Meaning: Provider hasn't finished processing

Causes:
├─ Normal (just needs more time) — Wait up to 20 minutes
├─ Blocked by fraud check — May take 30 min to hours
├─ Provider system slow — Rare but happens
└─ Network issue — Check provider status page

Action:
├─ Wait 20 minutes if recently created
├─ Check provider's status page
├─ Contact provider support if > 30 min pending
└─ Monitor webhook logs for completion
```

**If Provider Says: Completed (Status 100/SETTLED)**

```
Meaning: Provider finished, but our system missed the update

Investigation:
├─ Check webhook logs: Did we receive completion webhook?
│  - If YES: Why didn't we process it?
│  - If NO: Why didn't webhook arrive?
└─ Check our transaction status: Did we update it?

Cause: Webhook Failure
├─ Network interruption (our side)
├─ Webhook URL misconfigured
├─ Webhook handler crashed
└─ Request timeout

Fix:
├─ Manually mark transaction as complete
├─ Credit user's balance (should already have reserve)
├─ Send user notification: "✓ Payment completed"
├─ Investigate: Why webhook didn't trigger update
```

**If Provider Says: Failed (Status 3/FAILED)**

```
Meaning: Provider rejected the transaction

Investigate error:
├─ Check provider response: What was error?
│  Examples:
│  ├─ "Insufficient funds"
│  ├─ "Account locked"
│  ├─ "Daily limit exceeded"
│  └─ "Recipient account not found"
└─ Return error to user with explanation

Fix:
├─ Mark transaction as failed
├─ Return balance to sender (unreserve the amount)
├─ Show user error message from provider
└─ Suggest action (add funds, try again later, contact support)
```

---

### Problem 2: "Balance is wrong - shows £100 but should be £150"

**Symptoms:**
- User balance doesn't match their transactions
- "Missing" money somewhere
- Balance calculation error

### Investigation Checklist

**Step 1: Calculate expected balance**

```bash
# Sum all transactions for this user
SELECT 
  SUM(CASE WHEN type='receive' OR type='deposit' 
       THEN amount - COALESCE(provider_fee, 0)
       ELSE 0 END) as credits,
  SUM(CASE WHEN type='send' OR type='withdrawal'
       THEN amount + COALESCE(provider_fee, 0)
       ELSE 0 END) as debits
FROM transactions
WHERE user_id = 'user_...' AND status = 'completed';

Expected balance = credits - debits
Actual balance = [from database]
Difference = Expected - Actual
```

**Step 2: Investigate the discrepancy**

```
Difference = £50 (we have less than expected)

Check:
├─ Is there a pending transaction for £50? (Yes? Wait for it)
├─ Is there a failed transaction we show as completed? (Fix status)
├─ Did we miss recording a fee? (Add fee)
├─ Did a withdrawal fail but we show as sent? (Fix status)
└─ Database lag? (Wait 30 seconds, recalculate)
```

**Step 3: Check provider balance**

```bash
# Get authoritative balance from GateHub/PTI/Xago
Provider says: Alice has £100

Compare:
├─ Our calculated: £150
├─ Difference: £50 (we think we have more)
│
Who's right?
├─ Provider = Authoritative for real money
├─ Us = Authoritative if we're using internal transfers
└─ Action: Determine cause and reconcile
```

### Root Causes & Fixes

**Case 1: Unprocessed Deposits**

```
Scenario: Alice has pending deposit for £50

Our calculation:
├─ Only counts "completed" transactions
└─ Pending deposits not included (correct)

User perception:
├─ "I deposited £50, why don't I have it?"
└─ Doesn't understand "pending" status

Fix:
├─ Wait for deposit to complete
├─ If stuck > 20 min: Investigate with provider
└─ Educate user about pending status
```

**Case 2: Failed Withdrawal That We Marked as Complete**

```
Scenario: Withdrawal to bank failed but we show as completed

Transaction record:
├─ Status: "completed"
├─ Type: "withdrawal"
├─ Amount: £50
└─ But money never left our account

Fix:
├─ Query provider: "Did withdrawal succeed?"
├─ If NO: Update our transaction status to "failed"
├─ Recharge balance (add the £50 back)
├─ Notify user: "Withdrawal failed, trying again"
```

**Case 3: Fee Charged But Not Recorded**

```
Scenario: Provider charged fee we didn't track

Transaction:
├─ Amount: £100
├─ Provider Fee: £0 (should be £2!)
└─ Our balance change: -£100 (should be -£102)

Fix:
├─ Query provider for actual fee charged
├─ Update transaction record: fee = £2
├─ Recalculate: Balance should decrease additional £2
└─ If already final: Manual adjustment
```

**Case 4: Concurrent Transactions Causing Race Condition**

```
Scenario: Two payments at same time, order confusion

Alice's balance: £100
├─ Sends Bob £50
├─ Simultaneously sends Charlie £60

Timing issue:
├─ T+0: Payment 1 reserves £50
├─ T+1: Payment 2 reserves £60
├─ T+2: Both finalize (£110 total deducted from £100!)
└─ Result: Balance goes negative or transaction fails

Fix:
├─ Check: Did both complete or just one?
├─ Investigate: Transaction order in log
├─ Resolution: Likely one failed (insufficient funds)
├─ Recover: Mark second as failed, unreserve amount
```

---

### Problem 3: "Sender sent money but recipient didn't receive it"

**Symptoms:**
- Sender sees "sent" ✓
- Recipient sees nothing
- Money disappeared?

### Investigation Steps

**Step 1: Verify sender's transaction**

```bash
SELECT * FROM transactions
WHERE user_id = 'sender_...' 
  AND type = 'send'
  AND payment_id = 'pay_...'
  AND timestamp > NOW() - INTERVAL '1 hour';

Result:
├─ Alice's send transaction: Status = "completed" ✓
├─ Amount: £100
├─ Receiver: bob@example.com
└─ Provider confirmed: ✓
```

✓ Sender side is good, continue.

**Step 2: Check recipient's transaction**

```bash
SELECT * FROM transactions
WHERE user_id = 'receiver_...'
  AND type = 'receive'
  AND payment_id = 'pay_...'
  AND timestamp > NOW() - INTERVAL '1 hour';

Result:
├─ Bob's receive transaction: MISSING!
└─ No matching receive for the payment
```

✗ Bob has no matching receive transaction. This is the problem.

**Step 3: Query the payment record**

```bash
SELECT * FROM payments
WHERE id = 'pay_...';

Payment record:
├─ Sender: alice@example.com (user_123)
├─ Receiver: bob@example.com  (user_456)
├─ Status: completed
├─ Sender Transaction ID: txn_send_abc123 ✓
├─ Receiver Transaction ID: txn_receive_??? (MISSING!)
└─ Provider Response: Shows transfer succeeded
```

**Step 4: Check provider directly**

```bash
# Ask provider: Did money reach recipient's account?

GateHub: GET /core/v1/wallets/{bob_wallet_address}/balances

Result:
├─ GBP balance includes the £100
├─ Provider side: ✓ Money definitely arrived
└─ Conclusion: Provider saw it, Bob's balance updated there
```

### Root Causes & Fixes

**Case 1: Webhook Arrived But Receive Transaction Wasn't Created**

```
Timeline:
├─ Sender's send finalized
├─ Provider webhook: "Transfer to Bob completed"
├─ Our webhook handler: Received ✓
├─ Problem: Handler crashed before creating Bob's transaction
└─ Result: Bob's transaction missing, balance correct at provider

Fix:
├─ Manually create Bob's receive transaction
├─ Mark as completed
├─ Credit Bob's balance
├─ Notify Bob: "You received £100 from Alice" (now)
```

**Case 2: Webhook Never Arrived**

```
Timeline:
├─ Provider completed the transfer
├─ Provider webhook: Sent but lost (network issue)
├─ Our system: Still shows "pending"
├─ Result: Both ledgers out of sync

Discovery:
├─ Provider says: Money delivered (confirmed via API)
├─ Our ledger says: Still pending
└─ Time elapsed: > 20 minutes

Fix:
├─ Manually fetch status from provider API
├─ Get confirmation that transfer completed
├─ Create both transactions (send already exists)
├─ Create receive transaction for Bob
├─ Notify: Alice "✓ Sent" (already done)
├─ Notify: Bob "You received £100"
```

**Case 3: Different User Found During Lookup**

```
Scenario: Recipient email matches wrong user

Address resolution:
├─ System looks up: bob@example.com
├─ Finds: Wrong Bob (id_999 instead of id_456)
└─ Sends money to wrong person!

Prevention:
├─ Always verify identity match before transfer
├─ Use stable user IDs, not email addresses
├─ Warn user if recipient just signed up recently

Recovery:
├─ Locate actual funds (in wrong Bob's account)
├─ If wrong Bob hasn't spent it: Ask to return
├─ If spent: Escalate (financial loss)
├─ Apologize to user, refund from reserves
```

---

### Problem 4: "Our balance doesn't match provider's balance"

**Symptoms:**
- Our official balance: $10,000
- Provider's balance: $9,500
- Difference: $500 (we think we have more)

### Investigation Steps

**Step 1: Pull official balance**

```bash
# Fetch from provider
GateHub: GET /core/v1/wallets/{wallet_address}/balances
Response:
{
  "balances": [
    {
      "currency": "USD",
      "amount": "9500.00"
    }
  ]
}

Provider says: $9,500
```

**Step 2: Calculate our expected balance**

```bash
# Sum completed transactions
SELECT SUM(CASE 
  WHEN type IN ('deposit', 'receive') THEN amount - COALESCE(provider_fee, 0)
  WHEN type IN ('withdrawal', 'send') THEN -(amount + COALESCE(provider_fee, 0))
  ELSE 0
END) as expected_balance
FROM transactions
WHERE status = 'completed'
  AND user_id = 'user_...'
  AND currency = 'USD';

Our calculation: $10,000
```

**Step 3: Find the discrepancy**

```
Difference: Provider $9,500 vs Our $10,000 = $500 gap (we think we have more)

Question: What transaction might account for $500?

Investigate:
├─ Recent deposits? (Check pending deposits = $500?)
├─ Recent withdrawals? (Check failed withdrawal = $500?)
├─ Pending transaction? (Holding $500?)
├─ Fee charged we didn't record? (Opposite sign)
└─ Double-credit? (Recorded twice?)
```

**Step 4: Find the specific transaction**

```bash
# Look for $500 within last 24 hours
SELECT * FROM transactions
WHERE user_id = 'user_...' 
  AND ABS(amount) = 500
  AND timestamp > NOW() - INTERVAL '24 hours'
ORDER BY timestamp DESC;

Potential matches:
├─ Transaction ID: txn_456
├─ Type: deposit
├─ Amount: $500
├─ Status: "completed" (at our end)
├─ Provider confirmed: NEVER CHECKED
└─ Action: Query provider about this deposit
```

### Root Causes & Fixes

**Case 1: Deposit Completed at Provider But Transaction Shows Pending**

```
Transaction record:
├─ Type: deposit
├─ Amount: $500
├─ Status: "pending" (our system)
├─ Provider says: "completed" ✓

Investigation:
├─ Webhook never arrived (network issue)
├─ Reason: Interledger webhook handler was down
├─ Proof: Check webhook logs (no entry for this transaction)

Fix:
├─ Mark transaction as "completed"
├─ Confirm our expected balance now matches provider
├─ Action: Investigate webhook handler downtime
└─ Prevention: Set up alerts for webhook failures
```

**Case 2: Internal Transfer at Xago We Forgot to Reconcile**

```
Scenario: Xago provider, two users both at Xago

Our ledger:
├─ Alice -$500 (sent to Bob)
├─ Bob +$500 (received from Alice)
├─ Total: Balanced at our level

Provider's ledger (Xago):
├─ Doesn't know about internal transfer
├─ Our account balance: -$500 only
└─ Discrepancy expected for internal transfers

Reconciliation:
├─ Expected: Us +$500 ahead (because of internal transfer)
├─ Actual: Provider -$500 (doesn't track our internal transfer)
├─ Assessment: "Working as designed"
└─ Action: No fix needed, continue operating

When we settle with Xago:
├─ We report: "Internal transfer of $500 between our users"
├─ Xago confirms: "OK, we see the net effect in our balance"
└─ Settlement proceeds
```

**Case 3: Webhook Arrived, Transaction Marked Complete, But User Saw Failure**

```
Scenario: False negative to user

Timeline:
├─ T+0: Deposit initiated
├─ T+2: Our system: "Pending, let me tell provider"
├─ T+3: Provider sends webhook: "Completed"
├─ T+4: Our webhook handler: Updates transaction to complete ✓
├─ T+5: But user's app still shows "Processing"
│        (Frontend didn't refresh in time)
└─ T+10: User refreshes app: "✓ Deposit complete!"

Symptoms:
├─ User reports: "Money disappeared"
├─ Check system: Transaction is actually complete
├─ Provider confirms: Balance correct

Cause: Frontend display lag, not actual issue

Fix:
├─ Tell user to refresh browser
├─ Confirm: "Your deposit completed successfully"
├─ No system fix needed
```

---

## Temporal Debugging Guide

Temporal is the workflow orchestration engine that manages payment processing. When payments seem stuck or behave unexpectedly, Temporal's Web UI provides visibility into exactly what's happening.

**Accessing Temporal Web UI:**
- URL: `http://localhost:8233` (or your Temporal server address)
- No authentication required for local development
- Production: Requires proper credentials

### Understanding Payment Workflow Structure

Every payment creates **three workflows** that work together:

```mermaid
graph TD
    Payment["PaymentWorkflow<br/>ID: payments_&lt;uuid&gt;"]
    PayIn["PayinWorkflow<br/>ID: payment_pay_in_&lt;uuid&gt;"]
    PayOut["PayoutWorkflow<br/>ID: payment_pay_out_&lt;uuid&gt;"]
    
    Payment -->|starts| PayIn
    Payment -->|starts| PayOut
    PayIn -->|signals| PayOut
    
    PayIn -->|waits for| Webhook["Provider Webhook<br/>(transfer completed)"]
    Webhook -->|signals| PayIn
    PayIn -->|completes| PayOut
    
    style Payment fill:#e3f2fd,stroke:#1976d2
    style PayIn fill:#fff9c4,stroke:#f57f17
    style PayOut fill:#f3e5f5,stroke:#7b1fa2
    style Webhook fill:#e8f5e9,stroke:#388e3c
```

**Workflow IDs to search for:**
- **Parent workflow**: `payments_<payment_uuid>`
- **Pay-in workflow**: `payment_pay_in_<payment_uuid>`
- **Pay-out workflow**: `payment_pay_out_<payment_uuid>`

### Finding a Stuck Payment in Temporal

**Step 1: Get the Payment UUID**

From your database:
```sql
SELECT id, public_id, state, sender_id, receiver_id, created_at
FROM payments 
WHERE id = 'payment_uuid' 
  OR public_id = 'public_id_from_user';
```

Copy the `id` (UUID format like `57d4e5c4-957d-4638-847b-aad5e5714614`).

**Step 2: Search for the Workflow**

1. Open Temporal Web UI: `http://localhost:8233`
2. Go to **"Workflows"** tab (left sidebar)
3. In the search box, enter:
   - `payment_pay_in_<your_uuid>` for the sender side
   - `payment_pay_out_<your_uuid>` for the receiver side
   - `payments_<your_uuid>` for the parent orchestrator

**Step 3: Check Workflow Status**

Each workflow will show one of these statuses:

| Status | Meaning | Action |
|--------|---------|--------|
| **Running** | Workflow is actively processing or waiting | Check how long it's been running — see details below |
| **Completed** | Workflow finished successfully | Check execution time — should be <30 seconds normally |
| **Failed** | Workflow encountered an error | Check error message in workflow details |
| **Terminated** | Manually stopped or system shutdown | Investigate why it was terminated |
| **Timed Out** | Exceeded maximum allowed time | Check activity timeouts in workflow history |

### Investigating a Running Workflow

When a workflow shows "Running" status:

**Step 1: Click on the Workflow ID** to see details

**Step 2: Check the "Summary" section**
```
WorkflowId:      payment_pay_in_57d4e5c4-957d-4638-847b-aad5e5714614
Type:            PayinWorkflow
Status:          Running
Start Time:      3 minutes ago
Run Time:        3 minutes
```

**Questions to ask:**
- **How long has it been running?**
  - `< 5 minutes` → Normal, likely waiting for webhook
  - `5-20 minutes` → Monitor, should complete soon
  - `> 20 minutes` → Investigate why webhook hasn't arrived

**Step 3: View the "Pending Activities" section**

If the workflow is waiting, you'll see:
```
Pending Activities:
  (none) — Workflow is waiting for a signal
```

This means the workflow has finished all its tasks and is now waiting for the provider's webhook to arrive.

**Step 4: Check the "History" tab**

The history shows every step the workflow took. Look for:

1. **Activities that completed:**
   ```
   ActivityTaskScheduled: ReserveBalance
   ActivityTaskStarted:   ReserveBalance
   ActivityTaskCompleted: ReserveBalance
   ```
   ✓ This activity finished successfully

2. **The transfer call to the provider:**
   ```
   ActivityTaskScheduled: GatehubTransfer (or XagoTransfer, PTITransfer)
   ActivityTaskStarted:   GatehubTransfer
   ActivityTaskCompleted: GatehubTransfer
   ```
   ✓ Transfer was sent to provider

3. **Signal waiting (the important part):**
   ```
   [All activities complete]
   [No "WorkflowExecutionSignaled" event yet]
   ```
   ⏳ Workflow is waiting for provider webhook

4. **When webhook arrives:**
   ```
   WorkflowExecutionSignaled
   ```
   ✓ Webhook received!

5. **Post-webhook activities:**
   ```
   ActivityTaskScheduled: UpdatePayInTransactionState
   ActivityTaskCompleted: UpdatePayInTransactionState
   WorkflowExecutionCompleted
   ```
   ✓ Payment finalized

### Common Temporal Scenarios

#### Scenario 1: Workflow Waiting for Signal (Normal)

**What you see:**
```
Status:    Running
Run Time:  2 minutes
History:   - ReserveBalance ✓
           - GatehubTransfer ✓
           - SaveGatehubTransfer ✓
           - (waiting for signal)
```

**Meaning:** The workflow sent the payment to the provider and is now waiting for the webhook confirmation.

**Action:**
- **If < 5 minutes:** Wait, this is normal
- **If 5-20 minutes:** Check webhook logs to see if webhook arrived
- **If > 20 minutes:** Provider might be slow, or webhook lost. See [Webhook Troubleshooting](#webhook-troubleshooting-using-temporal)

####  Scenario 2: Workflow Stuck on an Activity

**What you see:**
```
Status:    Running
Run Time:  10 minutes
Pending Activities:
  - GatehubTransfer (started 10 minutes ago, retrying)
```

**Meaning:** An activity is stuck or failing repeatedly.

**Action:**
1. Click on the activity in History to see error details
2. Common causes:
   - Provider API is down (check provider status)
   - Network timeout (check connectivity)
   - Invalid credentials (check config)
   - Rate limit exceeded (wait and retry)

**To retry manually:**
- Workflows auto-retry failed activities
- If you need to force a retry, you can terminate and restart the workflow (last resort)

#### Scenario 3: Workflow Failed

**What you see:**
```
Status:     Failed
Run Time:   0.5 seconds
Error:      "insufficient funds"
```

**Meaning:** The workflow encountered a business logic error and stopped.

**Action:**
1. Read the error message in the workflow details
2. Common errors:
   - `"insufficient funds"` → User doesn't have enough money, check balance
   - `"account not found"` → Linked account missing, check user's accounts
   - `"provider error: ..."` → Provider rejected transaction, check provider response

**Resolution:**
- For business errors (insufficient funds, etc.): Fix the underlying issue and create a new payment
- For technical errors (timeout, network): Retry the payment

#### Scenario 4: Signal Received, Still Running

**What you see:**
```
Status:    Running
Run Time:  5 minutes
History:   - GatehubTransfer ✓
           - WorkflowExecutionSignaled ✓ (3 minutes ago)
           - UpdatePayInTransactionState (started 3 minutes ago, still running)
```

**Meaning:** Webhook arrived, but the finalization activity is stuck.

**Action:**
1. Check the activity details to see what's failing
2. Common causes:
   - Database is slow or locked
   - Pacioli ledger is unavailable
   - Network issue connecting to internal services
3. Check system health: database, Pacioli, network

### Webhook Troubleshooting Using Temporal

When a payment seems stuck waiting for a webhook:

**Step 1: Confirm Workflow is Waiting**

In Temporal UI, the workflow history should end with something like:
```
ActivityTaskCompleted: SaveGatehubTransfer
[no more events]
```

No `WorkflowExecutionSignaled` event means webhook hasn't arrived yet.

**Step 2: Check Provider Status**

Query the provider directly:
```bash
# For GateHub
curl https://mockgatehub.interledger.test/core/v1/transactions/<txn_id>

# Check status field
{
  "id": "txn_abc123",
  "status": 100,  ← Completed at provider
  "amount": "100.00"
}
```

**If provider says "completed" but no signal in Temporal:**

This means the webhook was lost or never sent.

**Step 3: Check Webhook Logs**

```bash
# Search backend logs for webhook receipts
docker compose logs backend | grep "Webhook received"

# Search for the specific payment ID
docker compose logs backend | grep "57d4e5c4-957d-4638-847b-aad5e5714614"
```

**If no webhook found in logs:**
- Network issue prevented webhook delivery
- Provider didn't send webhook (check provider logs)
- Webhook URL misconfigured

**Step 4: Manual Signal (Emergency Recovery)**

If the provider completed the transaction but webhook won't arrive, you can manually signal the workflow:

⚠️ **WARNING: Only do this if you're certain the provider confirmed completion!**

```bash
# Using Temporal CLI
docker compose exec temporal temporal workflow signal \
  --workflow-id "payment_pay_in_57d4e5c4-957d-4638-847b-aad5e5714614" \
  --name "payment_gatehub_signals" \
  --input '{"tx_uuid":"txn_abc123","status":"100"}'
```

Or use the Temporal Web UI:
1. Open the workflow
2. Click "Signal" button (top right)
3. Signal Name: `payment_gatehub_signals`
4. Signal Input:
   ```json
   {
     "tx_uuid": "txn_abc123",
     "status": "100"
   }
   ```
5. Click "Send Signal"

The workflow should immediately proceed with finalization.

### Temporal Best Practices for Support

1. **Always check Temporal first for "stuck" payments**
   - Faster than database queries
   - Shows exactly where the workflow is waiting
   - Reveals errors immediately

2. **Use the History tab to trace execution**
   - See which activities completed
   - See which activities failed
   - See timing between steps

3. **Look for signals to understand webhook delivery**
   - `WorkflowExecutionSignaled` = webhook arrived
   - Missing signal = webhook lost or delayed

4. **Don't terminate workflows unless absolutely necessary**
   - Workflows have built-in retry logic
   - Terminating loses state and may cause data inconsistency
   - Only terminate for unrecoverable errors

5. **Document workflow IDs when investigating**
   - Payment UUID → Workflow IDs mapping
   - Makes follow-up investigation easier
   - Helps with escalation to engineering

### Temporal Workflow Lifecycle Summary

```mermaid
stateDiagram-v2
    [*] --> Running: Payment initiated
    
    Running --> Activities: Execute business logic
    
    Activities --> Waiting: Sent to provider
    note right of Waiting: Workflow waits for webhook<br/>(no timeout, waits indefinitely)
    
    Waiting --> SignalReceived: Webhook arrives
    SignalReceived --> Finalization: Complete transaction
    
    Finalization --> Completed: Success
    Finalization --> Failed: Error during finalization
    
    Activities --> Failed: Activity error
    
    Completed --> [*]
    Failed --> [*]
```

**Key timing expectations:**
- Activities phase: 1-5 seconds (reserve balance, call provider)
- Waiting phase: 1-30 seconds normally (webhook arrival)
- Finalization phase: 1-3 seconds (update ledger, mark complete)
- **Total normal runtime: 5-40 seconds**
- **Stuck if: > 5 minutes with no signal**

---

## Escalation Logic

**When to investigate locally:**
- Recent transactions (< 1 hour)
- Clear user error (wrong amount, etc.)
- Webhook/network issues you can verify

**When to contact provider support:**
- Transaction status disputes (provider says different)
- Outage or service degradation
- API errors (500 responses, timeouts)
- Credentials not working

**When to escalate to finance team:**
- Settlement discrepancies
- Large balances missing
- Multiple users affected
- Fraud suspicion

---

## See Also

- [Payments & Transactions Guide](payments-explainer.md) — Overview and navigation hub
- [Ledger System Architecture](ledger-system-explainer.md) — How discrepancies form and resolve
- [Transaction Types Reference](transaction-types-explainer.md) — Transaction fields and states
- [Provider Payments Guide](provider-payments-guide.md) — Provider-specific behavior and limits

---

*Last updated: March 3, 2026*  
*Audience: Support staff, Operations, First-line debugging*
