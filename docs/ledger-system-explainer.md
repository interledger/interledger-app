# The Two-Ledger System: Pacioli & Provider Architecture

**Audience:** Operations managers, engineers debugging balance issues, finance staff  
**Time to read:** 15-20 minutes

> **Dual-ledger architecture reference.** Why we maintain two accounting systems and how they reconcile.

**Related documents:**
- [Payments Guide](payments-explainer.md) — Overview and navigation hub
- [Transaction Types Reference](transaction-types-explainer.md) — Transaction fields and lifecycle
- [Provider Payments Guide](provider-payments-guide.md) — Provider-specific ledger behavior
- [Payment Troubleshooting](payment-troubleshooting-guide.md) — Debugging balance discrepancies
- [Concepts](concepts.md) — Terminology (wallet, linked account, transaction)

**Quick Navigation:**
- **Why two ledgers?** → See [Why Have Two Ledgers](#why-have-two-ledgers-4-scenarios)
- **Balance mismatch between us and provider?** → See [When Ledgers Disagree](#when-ledgers-disagree)
- **How to reconcile?** → See [Reconciliation Process](#reconciliation-process)
- **Settlement procedures?** → See [Settlement Importance](#settlement-importance)
- **Real-world example?** → See [Real-World Example: Xago Transfer](#real-world-example-xago-transfer)

---

## Quick Summary

The Interledger App Wallet (a reference implementation for Account Servicing Entities integrating with the Interledger network) maintains **two separate accounting systems**:

1. **Our Ledger (Pacioli)** — immediate, optimistic, internal source of truth
2. **Provider's Ledger** — authoritative, external, slow but definitive

These ledgers must continuously reconcile to prevent money loss. Understanding how and why they exist is critical for operations.

> **Note:** This dual-ledger architecture is part of how the wallet demonstrates ASE integration patterns for the open payment network — balancing user experience (fast internal ledger) with regulatory/settlement requirements (provider ledgers).

---

## The Setup

### Our Ledger: Pacioli

- **Who owns it:** Interledger Foundation
- **How often updated:** Immediately (within milliseconds)
- **Source of truth for:** Our business logic, user experience, balance calculations
- **Can be changed:** Yes, retroactively for corrections

### Provider's Ledger

- **Who owns it:** GateHub, PTI, Xago, Chimoney (whichever provider)
- **How often updated:** When provider confirms, typically seconds to minutes
- **Source of truth for:** Legal/regulatory requirements, real money settlement
- **Can be changed:** No, we cannot modify it

---

## Why Have Two Ledgers?

Let's examine real scenarios where a single ledger would fail.

### Scenario 1: Network Failure During Payment

```
13:45:00 - Alice sends Bob £100
13:45:30 - Our system: Balance updated ✓
           (Pacioli ledger: Alice -£102, Bob +£100)
13:45:31 - Internet cuts out 🔌
13:47:00 - Internet comes back
13:48:00 - GateHub reply finally arrives: ✓ Confirmed
```

**If we ONLY relied on the provider:**
- We wouldn't know if the payment succeeded until GateHub responds (2+ minutes)
- Alice would see her balance unchanged while waiting
- Poor user experience; no way to show immediate feedback

**If we ONLY relied on our ledger:**
- We'd assume the payment worked and credit both accounts
- But GateHub might say "Transfer failed: Alice had insufficient funds"
- We'd create incorrect balances and have to reverse them
- Lost consistency and user trust

**With both ledgers:**
- Our ledger updates immediately → user sees "sent" status
- Provider ledger updates asynchronously → authoritative confirmation
- When provider responds, we verify our ledger matches
- Perfect UX without sacrificing financial accuracy

---

### Scenario 2: Provider System Temporarily Down

```
14:00:00 - Alice's bank account is temporarily locked
14:00:00 - Alice tries to withdraw £200 via PTI
14:01:00 - Our ledger: "Withdrawal created, let me tell PTI"
14:02:00 - PTI responds: "Sorry, account locked. Withdraw failed."
```

**If we ONLY had the provider ledger:**
- Alice would see nothing for 1-2 minutes
- No feedback, no indication what's happening
- Support gets flooded with "stuck withdrawal" tickets

**With both ledgers:**
- Alice immediately sees "Withdrawal processing"
- Backend handles PTI failure gracefully
- We can show helpful error: "Your bank temporarily locked your account"
- Better UX, lower support load

---

### Scenario 3: Internal Transfers (Xago/Same-Provider Case)

This is the most interesting case.

Some providers like Xago allow **internal transfers** — moving money between users without telling the provider.

```
Both Alice and Bob use Xago (South Africa)
Alice wants to send Bob R1,000

Option A (Tell Xago):
├─ We ask Xago: "Transfer R1,000 from Alice to Bob"
├─ Xago processes it
├─ Xago confirms completion
└─ Both ledgers agree

Option B (Internal Transfer):
├─ We check: "Alice and Bob both use Xago? ✓"
├─ Our Ledger (Pacioli): Deduct R1,000 from Alice, Add R1,000 to Bob ✓
├─ We DON'T tell Xago anything
├─ Xago's records remain unchanged
└─ But we know exactly who owns what
```

**Why use Option B (internal transfer)?**

- **Speed:** Instant (no API call to Xago)
- **Cost:** No per-transaction fee from Xago
- **Resilience:** Works if Xago's API is temporarily down
- **Simplicity:** Fewer external dependencies

**The catch:**

We must maintain our own ledger perfectly, because Xago's ledger won't reflect the transfer. 

If someone asks Xago "what's my balance?", Xago might say "R10,000" but we say "R9,000" (because Alice sent Bob R1,000). **Our answer is the real one for our users.**

This is why Pacioli (our ledger) is so critical: **for internal transfers, it's the only source of truth.**

**Reconciliation becomes essential:**

```
Daily:
├─ Pull official Xago balance: R10,000 (Xago doesn't know about our transfer)
├─ Calculate our expected: R9,000 (Alice -R1,000, Bob +R1,000)
├─ Understand: "Xago is behind us by R1,000"
├─ Action: None — this is expected for internal transfers
└─ Continue operating normally
```

---

## The Normal Flow

### Step-by-Step: Alice Sends Bob Money

```
T+0:00 - Alice clicks "Send Bob £100"
          ↓
T+0:01 - Our system checks: "Alice has £100? Yes ✓"
         Pacioli ledger: Reserve £100 from Alice
          ↓
T+0:02 - System creates two pending transactions
         (Alice's send, Bob's receive)
          ↓
T+0:03 - System says to GateHub: "Create transfer of £100"
          ↓
T+0:05 - GateHub responds in HTTP: "Created, status: pending"
         Backend records: Transaction ID = "txn_abc123"
          ↓
T+0:10 - GateHub webhook arrives: "Transfer completed!"
         Pacioli ledger: Finalize - Deduct £102 from Alice, Add £100 to Bob
          ↓
T+0:11 - Frontend shows: "✓ Sent! Bob received £100"
```

**At each step:**

| Time | Pacioli (Our Ledger) | GateHub (Provider) | User Sees |
|------|---------------------|-------------------|-----------|
| T+0:03 | Alice -£100 (reserved) | Nothing | "Sending..." |
| T+0:05 | Alice -£100 (reserved) | Pending | "Sending..." |
| T+0:10 | Alice -£102, Bob +£100 | Completed | "✓ Sent" |

---

## When Ledgers Disagree

Occasionally, our ledger and the provider's ledger don't match. Here's how to diagnose.

### Case 1: We Completed But Provider Still Pending

```
Our Ledger:
├─ Transaction 45: Status = "completed"
├─ Alice balance: -£100
└─ Bob balance: +£100

Provider's Ledger:
├─ Transaction 45: Status = "pending"
└─ (Nothing posted to either account yet)

Why did this happen?
├─ Option A: Webhook arrived, we marked it done, but provider webhook was delayed/lost
├─ Option B: Our clock is wrong
├─ Option C: Provider's API state is different from their webhook system
```

**Investigation:**

1. Check webhook logs: Did we receive confirmation from GateHub?
2. If yes: This is expected network lag. Pacioli is ahead, provider will catch up.
3. If no: Something unusual. Was there a provider outage?
4. Action: Wait 20 minutes and check. Provider eventually catches up or fails.

### Case 2: Provider Completed But We Still Pending

```
Our Ledger:
├─ Transaction 45: Status = "pending"
├─ Alice balance: unchanged (reserved)
└─ Bob balance: unchanged

Provider's Ledger:
├─ Transaction 45: Status = "completed"
├─ Money moved
├─ Both users show balance changes

Why did this happen?
├─ Option A: Webhook never arrived (network issue)
├─ Option B: Webhook corrupted or malformed
├─ Option C: Our webhook handler crashed
└─ Option D: Provider sent webhook but we never processed it
```

**Investigation:**

1. Check webhook logs: Is there a received webhook for txn_45?
2. If yes: Why didn't we process it? Check transaction handler logs.
3. If no: Provider completed but webhook never arrived. This is critical.
4. Action: Manually fetch transaction status from provider. If actually completed, update our ledger.

### Case 3: We're Ahead (Optimistic Lock)

```
Our Ledger:
├─ Alice sends Charlie £50
├─ Our Ledger: Deduct £50 from Alice, Add £50 to Charlie ✓

Provider's Ledger:
├─ Wallet Alice: Still shows old balance (hasn't processed yet)
└─ (Message to provider is still in flight or in queue)
```

**This is normal and expected.** Pacioli is optimistic; provider catches up in milliseconds.

### Case 4: Provider Ahead (Rare)

```
Scenario: Two transactions submitted at nearly same time, network timing varies

Our Ledger:
├─ Only first transaction confirmed
└─ Second still pending

Provider's Ledger:
├─ Both transactions completed
└─ Second one already processed
```

**Why rare:** Our Temporal workers usually process confirmations quickly enough that this doesn't happen.

**If it happens:** Not an error. Just means provider processed faster than our workers could update. Wait for next webhook.

---

## Reconciliation: Making Ledgers Agree

### What Reconciliation Means

Reconciliation is the **process of finding and fixing discrepancies** between our ledger and the provider's ledger.

```
Daily reconciliation:
├─ Pull official balance from provider: $10,000
├─ Calculate expected balance from our ledger: $9,500
├─ Difference: $500
│
Investigation:
├─ Find transaction T-234: $500 deposit
├─ Check: Did we credit the user? No!
├─ Root cause: Webhook for T-234 never arrived
│
Fix:
├─ Manually create transaction record for T-234
├─ Credit the user $500
└─ Recheck: Now both ledgers show $10,000 ✓
```

### The Reconciliation Process

```
Daily (Automated):
├─ Fetch provider's reported balance
├─ Compare to our calculated balance
├─ If match: Log "OK"
├─ If differ: Alert operations team

Weekly (Manual Review):
├─ Export transaction log from both systems
├─ Line-item comparison
├─ Identify missing transactions
├─ Create correction transactions as needed
└─ Document findings

Monthly (Financial):
├─ Create account statement for provider relationship
├─ Invoice calculations based on reconciled balances
├─ Prepare for cash settlement
```

### Common Reconciliation Issues

**Missing Transaction (Most Common)**

```
Provider shows: $10,000 balance
We expect: $9,500
Difference: +$500

Investigation:
├─ Provider's transaction log shows: Deposit of $500 (COMPLETED)
├─ Our transaction log shows: Nothing
├─ Cause: Webhook never delivered
├─ Fix: Create matching transaction in our system
└─ Result: Ledgers now match
```

**Incorrect Fee**

```
Provider says: "Transfer fee was $2.50"
We recorded: "$0" (testing with fees disabled)

Investigation:
├─ Check fee configuration
├─ Cause: Mock provider used in testing
├─ Fix: Enable real fees or update mock config
└─ Result: Ledgers now match
```

**Timing Misalignment**

```
Provider time: March 3, 14:05:00 UTC
Our time: March 3, 14:04:55 UTC

Network timing variance between systems can cause:
├─ Same transaction recorded in different fiscal periods
├─ Out-of-order completion marks
├─ Timestamp disagreements

Solution:
├─ Synchronize time across systems
├─ Use provider's timestamp as authoritative
└─ Resync local records if needed
```

---

## Why This Matters for Settlement

Settlement (the legal/financial reconciliation with providers) depends entirely on accurate ledger matching.

```
Monthly Settlement:
├─ We tell provider: "You owe us $X for completed transactions"
├─ Provider agrees: "Yes, records match"
├─ Agreement signed
├─ Payment processes
│
But if ledgers don't match earlier:
├─ We claim: "You owe us $10,000"
├─ Provider says: "Our records show $9,500"
├─ Dispute initiated 😢
├─ Payment delayed
├─ Relationship strained
```

This is why continuous reconciliation matters: **catch discrepancies early, fix them before settlement.**

---

## Real-World Example: Xago Settlement

Let's trace through an actual Xago reconciliation scenario to see all concepts together.

```
Week 1 (March 1 - March 7):

User Transactions:
├─ Alice sends Bob R1,000 (internal transfer)
├─ Charlie deposits R500
├─ Diana withdraws R200
└─ Edward receives R300 from external source (webhook)

Our Ledger (Pacioli):
├─ Total user balances: R1,600
├─ Transaction log: 4 entries
└─ Status: All completed

Xago's Ledger:
├─ Our account balance: R1,200 (doesn't know about internal Alice→Bob transfer)
├─ Completed deposits: R500
├─ Completed withdrawals: -R200
└─ Status: All confirmed

Reconciliation:
├─ We report: "Our users hold R1,600"
├─ Xago confirms: "We have received R1,200 from you"
├─ Difference: R400
│
Analysis:
├─ Internal transfer (Alice→Bob): R1,000 (Xago doesn't track these) ✓ Expected
├─ Net external activity (Charlie +500, Diana -200, Edward +300): +R600
├─ Wait, math doesn't match!
├─ Action: Investigate Edward's deposit
│
Investigation:
├─ Edward's deposit: "Received R300" (our ledger shows completed)
├─ Check Xago: No webhook for Edward
├─ Action: Fetch Edward's transaction status from Xago API
│
Finding:
├─ Xago says: "Edward deposit is PENDING, not completed"
├─ Our ledger says: "Completed"
├─ Root cause: Webhook didn't arrive for Edward
│
Fix:
├─ Update our ledger: Mark Edward's transaction as "pending" or "failed"
├─ Investigate why webhook didn't arrive
├─ Set up manual retry or escalation
│
Retry:
├─ Edward retries deposit
├─ Check webhook arrives this time
├─ Mark as completed
│
Final Reconciliation:
├─ Our users hold: R1,600 (internal: 1000, external: 600)
├─ Xago has received: R1,200 (1000 internal, 200 net external activity)
├─ Match! ✓
└─ Settlement proceeds
```

---

## Key Concepts

### Optimism vs Authoritativenss

| System | Optimism | Speed | Reliability | Authority |
|--------|----------|-------|-------------|-----------|
| Pacioli (Ours) | High | Immediate | Good | Our operations |
| Provider | Low | Delayed | High | Legal/financial truth |

**How we balance them:**
- Be optimistic with Pacioli (immediate feedback)
- Be thorough with provider (verify everything)
- Reconcile frequently (catch errors early)

### The 20-Minute Rule

Most providers resolve transaction status within 20 minutes:

```
T+0:00   - Transaction created
T+0:05   - Webhook arrives (expected)
T+20:00  - If no webhook: Manual poll from provider
T+20:30  - Mark completed/failed based on poll result
```

If nothing resolved by T+20:00, something went wrong. Escalate.

### Settlement Periods

Different providers have different settlement windows:

| Provider | Settlement | Frequency |
|----------|-----------|-----------|
| GateHub | Continuous | Real-time |
| PTI | Batch | Daily |
| Xago | Daily | Nightly |
| Chimoney | Per-transfer or weekly | Varies |

**Plan your reconciliation accordingly.**

---

## Troubleshooting Guide for Ledger Issues

### "Our balance doesn't match provider's"

**Step 1:** Identify the difference
```bash
Our balance: $10,000
Provider balance: $9,500
Difference: $500 (we think we have more)
```

**Step 2:** Find the transaction
```
Which transactions might cause $500 difference?
├─ Deposits we record but provider hasn't confirmed? (Check pending status)
├─ Fees we subtracted but provider says they didn't charge?
├─ Withdrawals we show as completed but provider shows as pending?
└─ Internal transfers we created but provider doesn't know about?
```

**Step 3:** Investigate root cause
```
IF internal transfer:
  ├─ Expected (provider doesn't track these)
  └─ Continue

IF pending transaction:
  ├─ Wait 20 minutes
  ├─ Check webhook logs
  ├─ Manually fetch status
  └─ Update our ledger accordingly

IF fee mismatch:
  ├─ Check fee configuration
  ├─ Check provider's fee response
  ├─ Correct our calculation
  └─ Update transaction
```

**Step 4:** Reconcile
```
After fix:
├─ Recalculate our balance
├─ Confirm it matches provider
└─ Document the issue
```

### "Webhook arrived but transaction still pending"

**Likely cause:** Webhook handler crash or database write failure

**Investigation:**
```
1. Check webhook log: Was it received?
2. If yes: Was it processed? (Check success flag)
3. If not: Why? Check handler logs for errors
4. Action: Fix the error and reprocess webhook
```

### "Provider says completed but we show pending"

**Likely cause:** Our polling/verification failed

**Investigation:**
```
1. When did we last check provider status?
2. Did we get a response?
3. If yes: Why didn't we update our ledger?
4. Action: Manually fetch, update our records
```

---

## See Also

- [Payments & Transactions Guide](payments-explainer.md) — Overview and navigation hub
- [Transaction Types Reference](transaction-types-explainer.md) — What fields transactions contain
- [Payment Troubleshooting Guide](payment-troubleshooting-guide.md) — Debugging common payment issues
- [Provider Payments Guide](provider-payments-guide.md) — Provider-specific differences

---

*Last updated: March 3, 2026*  
*Audience: Operations managers, Finance staff, Engineers*
