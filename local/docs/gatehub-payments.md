# GateHub Payments Workflow

## Overview

This document describes the payment workflow in the Interledger App Wallet, specifically focusing on peer-to-peer (P2P) payments using GateHub as a provider. It covers the complete journey of a payment from sender to receiver, including database interactions, Temporal workflow orchestration, webhook handling, and troubleshooting guidance.

## Key Architectural Components

### Transaction Status Codes

GateHub API uses specific integer status codes for transaction states:

| Status Code | Constant Name | Meaning | Next Action |
|------------|---------------|---------|-------------|
| `1` | `TransactionStatusPending` | Transaction created, processing | Wait for completion |
| `100` | `TransactionStatusCompleted` | Transaction successfully completed | Mark payment complete |
| `3` | `TransactionStatusFailed` | Transaction failed | Mark payment failed |

**Critical**: The PayOut workflow polls or receives signals about transaction completion, but only exits the polling loop when `status == 100`. Using incorrect status codes (e.g., `2` instead of `100`) will cause workflows to poll indefinitely.

### Webhook and Signal Integration

Payments use a hybrid completion detection mechanism:
- **Primary**: Webhook notifications from GateHub/MockGatehub trigger Temporal signals
- **Fallback**: 20-minute polling timer checks transaction status if webhooks fail

This ensures reliable payment completion even with intermittent webhook delivery.

## Payment Types

The wallet supports several payment types:
- **Peer2Peer (P2P)**: Direct transfer between two wallet users
- **Withdrawal**: Transfer from wallet to an external bank account
- **Deposit**: Transfer from an external bank account to wallet
- **RafikiPeer2Peer**: P2P via the Rafiki open payments network
- **WebMonetization**: Incoming payments via web monetization
- **Rafiki2External**: Rafiki incoming payment to external account

## Payment Workflow Architecture

### High-Level Flow

```
1. Create Payment Request
   ↓
2. Validate & Identify Accounts
   ↓
3. Launch Parallel Workflows
   ├─ PayIn Workflow (sender side)
   │  ├─ Reserve balance
   │  ├─ Execute transfer (if needed)
   │  └─ Update transaction
   │
   └─ PayOut Workflow (receiver side)
      ├─ Wait for receiver to be ready
      ├─ Finalize sender's balance
      ├─ Assign balance to receiver
      └─ Create receiver transaction
   ↓
4. Verify Payment Success
   ↓
5. Send Notifications
```

### Detailed Sequence Diagrams

#### 1. Peer-to-Peer Payment Sequence

```
┌──────────────────────────────────────────────────────────────────────────────────┐
│                          PEER-TO-PEER PAYMENT FLOW                              │
└──────────────────────────────────────────────────────────────────────────────────┘

Sender App      Wallet Backend      Temporal Workflow     Database        Receiver
   │                │                      │                  │               │
   │─ Create Payment →                     │                  │               │
   │                │                      │                  │               │
   │                │─ Validate Payment ───→                  │               │
   │                │                      │                  │               │
   │                │  Set State: Processing                  │               │
   │                │                      │─ Store ──────────→               │
   │                │                      │                  │               │
   │                │  Launch PaymentWorkflow                 │               │
   │                │                      │                  │               │
   │                ├──── Parallel Workflows ────────────────┐                │
   │                │                      │                 │                │
   ┌────────────────┼──────────────────────┼─────┐           │                │
   │ PayIn Workflow │                      │     │           │                │
   │                │                      │     │           │                │
   │ 1. Check if pull needed               │     │           │                │
   │    (Yes for P2P)                      │     │           │                │
   │                │                      │     │           │                │
   │ 2. Lookup sender account              │     │           │                │
   │    (GateHub linked account)           │     │           │                │
   │                │──────────────────────────────────────→  │                │
   │                │    Read sender account balance          │                │
   │                │← ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─  │                │
   │                │                      │     │           │                │
   │ 3. Reserve Balance                    │     │           │                │
   │    (Lock funds in Pacioli ledger)     │     │           │                │
   │    ⚠️  NOTE: Updates internal Pacioli ledger only      │                │
   │        Does NOT call mockgatehub API                    │                │
   │        Mockgatehub balance unchanged at this point      │                │
   │                │──────────────────────────────────────→  │                │
   │                │    Pacioli.CreateTransfers(pending)     │                │
   │                │    (via GateHub.ReserveBalance)         │                │
   │                │← ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─  │                │
   │                │                      │     │           │                │
   │ 4. Create Send Transaction            │     │           │                │
   │                │─ CreateTransaction ──→    │           │                │
   │                │                      │     │           │                │
   │ 5. Add Pay-In Transfer                │     │           │                │
   │    (Debit balance)                    │     │           │                │
   │                │─ AddPayInTransfer ───→    │           │                │
   │                │                      │     │           │                │
   │ 6. Signal PayOut Workflow             │     │           │                │
   │                │─ Signal PayOut ──────→    │           │                │
   │                │                      │     │           │                │
   │ 7. Wait for GateHub Transaction       │     │           │                │
   │    ⚠️  HYBRID COMPLETION DETECTION:                    │                │
   │    - Webhook: Receives signal on 'payment_gatehub_signals' channel    │
   │    - Polling: Checks status every 20 minutes via GetGatehubTransfer   │
   │    - Selector: Waits for EITHER signal OR timer                       │
   │    - Status Check: Queries transaction.status after wake-up           │
   │    - Loop Exit: Breaks when status == 100 (completed)                 │
   │                │                      │     │           │                │
   └────────────────┴──────────────────────┴─────┘           │                │
   │                │                      │                 │                │
   ┌────────────────┼──────────────────────┼─────────────────┐                │
   │ PayOut Workflow│                      │                 │                │
   │                │                      │                 │                │
   │ 1. Await Signal from PayIn            │                 │                │
   │                │                      │                 │                │
   │ 2. Check Receiver Ready               │                 │                │
   │    (Has receive account)              │                 │                │
   │                │     Lookup default receive account     │                │
   │                │──────────────────────────────────────→  │                │
   │                │    Find GateHub account for currency    │                │
   │                │← ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─  │                │
   │                │                      │                 │                │
   │ 3. Finalize Sender Balance            │                 │                │
   │    (Commit reservation in Pacioli)    │                 │                 │
   │                │──────────────────────────────────────→  │                │
   │                │    Pacioli.CreateTransfers(post)        │                │
   │                │    (via GateHub.FinaliseReserve)        │                │
   │                │← ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─  │                │
   │                │                      │                 │                │
   │ 4. Assign Balance to Receiver         │                 │    CRITICAL ⚠  │
   │    (Credit via GateHub API call)      │                 │                │
   │    ⚠️  This DOES call mockgatehub /core/v1/transactions │                │
   │        Mockgatehub balance updated here                 │                │
   │                │──────────────────────────────────────────────────────→  │
   │                │    GateHub.CreateTransfer(hosted)       │                │
   │                │    → POST /core/v1/transactions         │                │
   │                │← ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─  ─ ─ ─ ─ ─ ─ ─  │
   │                │                      │                 │                │
   │ 5. Create Receive Transaction         │                 │                │
   │    CreatePayoutTransaction()          │                 │                │
   │                │─────────────────────→ New TX with      │                │
   │                │    CreateTransaction │ Status: PENDING │                │
   │                │                      │─ Store ────────→ │                │
   │                │                      │                 │                │
   │ 6. Set Receive TX to COMPLETED        │                 │                │
   │                │─────────────────────→ SetTransactionState              │
   │                │                      │─ Update ──────→ │                │
   │                │                      │                 │                │
   └────────────────┴──────────────────────┴─────────────────┘                │
   │                │                      │                 │                │
   │                │  Check Payment Success                │                │
   │                │                      │                 │                │
   │                │  Verify both TX completed              │                │
   │                │  Set Payment: COMPLETED               │                │
   │                │                      │─ Update ──────→ │                │
   │                │                      │                 │                │
   │← ─ Send Success Response ─            │                 │                │
   │                │                      │                 │                │
   │                │     Send Emails      │                 │                │
   │                │ (Sent & Received)    │                 │                │
   └────────────────────────────────────────────────────────────────────────────┘
```

#### 2. Account Setup and Balance Flow

```
┌─────────────────────────────────────────────────────────────┐
│           GATEHUB BALANCE MANAGEMENT                        │
└─────────────────────────────────────────────────────────────┘

Sender (dabatla@gmail.com)
  │
  ├─ Wallet: [wallet_id_1]
  │
  ├─ Linked Account (Send): GateHub USD
  │   ├─ Provider: GateHub
  │   ├─ Type: Balance (not Card/Bank)
  │   ├─ Send Currency: USD
  │   └─ Balance: 1,000 USD (in GateHub)
  │
  └─ Send Transaction [tx_send_1]
      ├─ State: PROCESSING → COMPLETED
      ├─ Type: Received (from sender perspective)
      └─ Transfers:
          └─ Debit Balance: -100 USD

Receiver (alice@example.com)
  │
  ├─ Wallet: [wallet_id_2]
  │
  ├─ Linked Account (Receive): GateHub USD
  │   ├─ Provider: GateHub
  │   ├─ Type: Balance (not Card/Bank)
  │   ├─ Receive Currency: USD
  │   └─ Balance: 500 USD (in GateHub)
  │       └─ +100 USD from payment → 600 USD
  │
  └─ Receive Transaction [tx_receive_1]
      ├─ State: PENDING → COMPLETED
      ├─ Type: Received
      └─ Transfers:
          └─ Credit Balance: +100 USD

Balance State Machine:
  Initial (Sender)    Initial (Receiver)
      │                     │
      ├─ Reserve ─────→ (unchanged)
      │  (Pacioli only)   (mockgatehub: 100 EUR)
      │  Total: 100       
      │  Available: 90    
      │  Pending: 10      
      │
      ├─ Finalize ─────→ Assign (mockgatehub API call)
      │    │               │
      │  (Pacioli)      (POST /core/v1/transactions)
      │  Post transfer   Sender: 100→90, Receiver: 0→10
      │                     │
      └────┴─→ ← ─ ─ ─ ─ ─┘
          Both completed = Money moved in mockgatehub!
```

## Database Schema

### Key Tables

#### payments
Stores payment records with state, sender/receiver, amounts, and accounts
```sql
CREATE TABLE payments (
  id UUID PRIMARY KEY,
  public_id TEXT UNIQUE,           -- Human-readable ID
  state INTEGER,                    -- 1=Created, 2=Processing, 3=Completed, 4=Failed
  sender_id TEXT,                   -- Sender identifier (wallet ID, email, etc)
  sender_id_type TEXT,              -- Type: wallet_id, email, etc
  sender_amount BIGINT,             -- Amount in smallest currency unit
  sender_currency TEXT,             -- e.g., USD, EUR, ZAR
  sender_account TEXT,              -- Linked account ID (GateHub account)
  send_transaction_id TEXT,         -- Transaction ID on sender side
  receiver_id TEXT,                 -- Receiver identifier
  receiver_id_type TEXT,            -- Type of receiver ID
  receiver_amount BIGINT,
  receiver_currency TEXT,
  receiver_account TEXT,            -- Linked account ID (GateHub account) ⚠️
  receive_transaction_id TEXT,      -- Transaction ID on receiver side
  type INTEGER,                      -- 1=P2P, 2=Withdrawal, 3=Deposit, etc
  note TEXT,
  created_at TIMESTAMP,
  updated_at TIMESTAMP,
  revision INTEGER
);
```

#### transactions
Stores individual transactions (send/receive) with state and transfers
```sql
CREATE TABLE transactions (
  id UUID PRIMARY KEY,
  wallet_id UUID,                   -- Owner of the transaction
  foreign_id TEXT,                  -- Links to payment ID
  type TEXT,                         -- Received, Sent, Withdrawal, etc
  state TEXT,                        -- pending, completed, failed
  provider TEXT,                     -- payments_engine, gatehub, xago, etc
  source TEXT,                       -- Wallet address of sender
  destination TEXT,                 -- Wallet address of receiver
  amount BIGINT,
  asset_code TEXT,
  asset_scale BIGINT,
  created_at TIMESTAMP,
  updated_at TIMESTAMP
);
```

#### transfers
Individual line items in a transaction (debit/credit operations)
```sql
CREATE TABLE transfers (
  id UUID PRIMARY KEY,
  transaction_id UUID,              -- Links to transaction
  linked_acc_id UUID,               -- Which GateHub account affected
  foreign_id TEXT,                  -- Links to payment ID
  type TEXT,                         -- DebitBalance, CreditBalance, etc
  state TEXT,                        -- pending, completed
  amount BIGINT,
  asset_code TEXT,
  asset_scale BIGINT,
  created_at TIMESTAMP,
  updated_at TIMESTAMP
);
```

#### linked_accounts
Maps wallet users to their external provider accounts (GateHub, Xago, etc)
```sql
CREATE TABLE linked_accounts (
  id UUID PRIMARY KEY,
  wallet_id UUID,
  provider TEXT,                    -- gatehub, xago, pti, chimoney
  type TEXT,                         -- balance, bank, card
  provider_id TEXT,                 -- External ID in the provider system
  send_currency TEXT,               -- What currency can be sent from this account
  receive_currency TEXT,            -- What currency can be received into this account
  created_at TIMESTAMP,
  updated_at TIMESTAMP
);
```

#### gatehub_users
Maps wallet users to their GateHub accounts
```sql
CREATE TABLE gatehub_users (
  id UUID PRIMARY KEY,
  external_id TEXT,                 -- GateHub user ID
  wallet_id UUID,
  external_customer_id TEXT,        -- GateHub customer ID
  external_account_id TEXT,         -- GateHub account ID
  created_at TIMESTAMP,
  updated_at TIMESTAMP
);
```

### Database Query Examples

#### 1. Check Payment Status
```sql
-- Get payment details
SELECT 
  id, public_id, state, 
  sender_id, receiver_id,
  sender_amount, sender_currency,
  receiver_amount, receiver_currency,
  sender_account, receiver_account,
  send_transaction_id, receive_transaction_id,
  created_at, updated_at
FROM payments
WHERE public_id = '12345-67890'
ORDER BY updated_at DESC;

-- Payment states:
-- 1 = Created
-- 2 = Processing
-- 3 = Completed (SUCCESS)
-- 4 = Failed
```

#### 2. Check Sender's Transaction
```sql
-- Get sender's transaction
SELECT t.id, t.wallet_id, t.state, t.type, t.amount
FROM transactions t
WHERE t.id = (SELECT send_transaction_id FROM payments WHERE id = 'payment-uuid')
  AND t.wallet_id = 'sender-wallet-id';

-- Check transfers (debit entries)
SELECT tf.id, tf.type, tf.state, tf.amount, tf.linked_acc_id
FROM transfers tf
WHERE tf.transaction_id = 'transaction-uuid'
ORDER BY tf.created_at;
```

#### 3. Check Receiver's Transaction ⚠️ CRITICAL
```sql
-- Get receiver's transaction
SELECT t.id, t.wallet_id, t.state, t.type, t.amount
FROM transactions t
WHERE t.id = (SELECT receive_transaction_id FROM payments WHERE id = 'payment-uuid')
  AND t.wallet_id = 'receiver-wallet-id';

-- This should show:
-- - State: 'completed'
-- - Type: 'Received'
-- - Transfers: Credit balance entries

-- If receive_transaction_id is NULL or transaction doesn't exist, 
-- money won't show up for receiver!
```

#### 4. Check Linked Accounts
```sql
-- Find GateHub accounts for a wallet
SELECT id, wallet_id, provider, type, provider_id, send_currency, receive_currency
FROM linked_accounts
WHERE wallet_id = 'wallet-uuid'
  AND provider = 'gatehub'
ORDER BY receive_currency;

-- Each currency should have:
-- - One account for sending (send_currency = 'USD')
-- - One account for receiving (receive_currency = 'USD')
```

#### 5. Full Payment Trace
```sql
-- Complete payment journey for debugging
WITH payment_details AS (
  SELECT 
    id as payment_id,
    public_id,
    state as payment_state,
    sender_id, receiver_id,
    sender_account, receiver_account,
    send_transaction_id, receive_transaction_id
  FROM payments
  WHERE id = 'payment-uuid'
)
SELECT 
  p.payment_id,
  p.public_id,
  p.payment_state,
  'SENDER' as perspective,
  t.id as transaction_id,
  t.state as tx_state,
  t.wallet_id,
  tf.type as transfer_type,
  tf.state as transfer_state,
  tf.amount
FROM payment_details p
LEFT JOIN transactions t ON t.id = p.send_transaction_id
LEFT JOIN transfers tf ON tf.transaction_id = t.id

UNION ALL

SELECT 
  p.payment_id,
  p.public_id,
  p.payment_state,
  'RECEIVER' as perspective,
  t.id as transaction_id,
  t.state as tx_state,
  t.wallet_id,
  tf.type as transfer_type,
  tf.state as transfer_state,
  tf.amount
FROM payment_details p
LEFT JOIN transactions t ON t.id = p.receive_transaction_id
LEFT JOIN transfers tf ON tf.transaction_id = t.id
ORDER BY perspective DESC;
```

## Checking MockGatehub Account Status

MockGatehub maintains account balances and tracks all transactions. Here's how to check status:

### 1. Check User Balance

```bash
curl http://localhost:8080/wallets/{user_id}/balance
```

Response:
```json
[
  {
    "currency": "USD",
    "vault_uuid": "450d2156-132a-4d3f-88c5-74822547658d",
    "balance": 1000.00
  }
]
```

### 2. Create Test Payment Request

```bash
# Create a GateHub payment request to send money
curl -X POST http://localhost:8080/core/v1/wallets \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "dabatla@gmail.com",
    "name": "Test Wallet",
    "type": "WALLET",
    "network": "xrpl"
  }'
```

### 3. List Transactions in MockGatehub

```bash
# Check transaction history
curl http://localhost:8080/core/v1/transactions?user_id=alice@example.com

# Response shows all transactions for the user
```

### 4. Verify Account Existence

```bash
# Get user details
curl http://localhost:8080/id/v1/users/{user_id}

# Response shows user KYC status, vaults, etc
```

### 5. Check Vault Status

```bash
# Each currency has a vault
# USD Vault: 450d2156-132a-4d3f-88c5-74822547658d
# EUR Vault: a09a0a2c-1a3a-44c5-a1b9-603a6eea9341

# Balances are stored by currency in MockGatehub
# GetBalance returns all 11 currencies even if zero
```

### 6. Using Local Development Tools

```bash
# If local dev environment is running, you can also:

# Check wallet transactions via GraphQL
# See wallet frontend at http://localhost:4003

# View database directly (if postgres available):
psql wallet_backend -c "SELECT * FROM payments ORDER BY updated_at DESC LIMIT 5;"

# View GateHub mock logs:
docker logs mockgatehub
```

## Common Issues and Debugging

### Issue 1: Payment Created But Not Showing for Receiver (FIXED)

**Symptoms:**
- Sender sees transaction completed
- Receiver sees nothing
- `receive_transaction_id` is NULL in database
- Sender transaction in "Pending" state (not completed)

**Root Cause:**
For Rafiki payments (where receiver is identified by an external wallet URL like `https://local.ilp.link/alice`), the `CreatePayoutTransaction()` activity was trying to look up the receiver as a local wallet, which failed. This prevented the receive transaction from being created and left the payment in an incomplete state.

**Solution (Fixed in commit):**
Skip `CreatePayoutTransaction` for external wallet URLs. For these payments, the receive transaction is created by the Rafiki webhook handler (`immediatePayment`) when the incoming payment notification arrives from Rafiki.

**How to Check if Fixed:**
```sql
-- After fix, payments with external receiver URLs should have:
-- 1. send_transaction_id populated
-- 2. receive_transaction_id populated (set by Rafiki webhook)
-- 3. Both transactions in 'completed' state
-- 4. Payment state = 3 (completed)

SELECT 
  id, public_id, state, 
  receiver_id,
  send_transaction_id, receive_transaction_id
FROM payments 
WHERE receiver_id LIKE 'https://%'
ORDER BY updated_at DESC LIMIT 5;
```

### Issue 2: Balance Assigned But Transaction Not Completed

**Symptoms:**
- Balance shows in MockGatehub
- Database transaction exists but state = 'pending'
- Receiver's app doesn't show money

**Root Cause:**
`UpdatePayoutTransactionState()` not called or failed

**How to Check:**
```sql
SELECT state FROM transactions WHERE id = 'receive-tx-uuid';
-- Should be 'completed', not 'pending'
```

### Issue 3: Receiver Account Not Found During Payment

**Symptoms:**
- Payment fails with error
- Logs show: "failed to find receiver account for currency"

**Root Cause:**
Receiver doesn't have linked account for the payment currency

**Solution:**
Ensure receiver has completed KYC and has linked GateHub account for USD/EUR/etc

### Issue 4: Payment Workflow Not Completing After External Transaction

**Symptoms:**
- PayOut workflow enters selector loop waiting for transaction completion
- Workflow remains RUNNING with 20-minute timer active
- GateHub transaction exists in mockgatehub but workflow doesn't detect completion
- Webhook is sent and received successfully, but workflow continues polling

**Root Cause:**
The PayOut workflow uses a selector pattern that waits for either:
1. A signal from the webhook handler (`payment_gatehub_signals` channel), OR
2. A 20-minute polling timer to check transaction status manually

After the selector receives either event, it queries the transaction status via `GetGatehubTransfer` activity and checks if `status == 100` (completed). If the status doesn't match, it loops back to waiting.

**Critical Transaction Status Codes:**
GateHub API uses specific status codes for transactions:
- `1` = Pending (transaction created, processing)
- `100` = Completed (transaction successfully completed)
- `3` = Failed (transaction failed)

The backend expects `status=100` to mark a transaction as complete. If mockgatehub or the external provider returns a different value (e.g., `status=2`), the workflow will continue polling indefinitely.

**How to Check:**
```sql
-- Check gatehub_transactions table for the external transaction ID
SELECT * FROM gatehub_transactions WHERE payment_id = 'payment-uuid';

-- Check transaction status in mockgatehub Redis (DB 2)
docker exec local-redis-1 redis-cli -n 2 GET "tx:<external_id>" | jq '.status'
-- Should return 100 for completed transactions
```

**Verify Webhook and Signal Delivery:**
```bash
# Check backend logs for webhook receipt and signal delivery
docker compose logs backend | grep -E "(Webhook received|SignalGatehubTransferComplete|SIGNAL RECEIVED)"

# Should show sequence:
# 1. Webhook received: event_type=core.deposit.completed
# 2. SignalGatehubTransferComplete called: external_transaction_id=<uuid>
# 3. Found payment for signal: payment_id=<uuid>
# 4. Successfully signaled workflow
# 5. ***SIGNAL RECEIVED*** on gatehub notify channel

# Check if workflow breaks from loop
docker compose logs backend | grep "Transaction completed - breaking polling loop"
```

**Troubleshooting Steps:**

1. **Verify transaction status in mockgatehub:**
   ```bash
   docker exec local-redis-1 redis-cli -n 2 GET "tx:<external_id>" | jq '.status'
   ```
   Expected: `100` (not `1`, `2`, or other values)

2. **Check webhook configuration:**
   ```bash
   # Ensure webhook URL and secret are configured
   docker compose exec mockgatehub env | grep WEBHOOK
   ```
   Expected: `WEBHOOK_URL=http://backend:8080/webhooks/gatehub`

3. **Verify webhook delivery:**
   ```bash
   # Check mockgatehub logs for webhook sending
   docker compose logs mockgatehub | grep -i webhook
   ```
   Expected: Status 200 responses from backend

4. **Trace workflow execution:**
   ```bash
   # Check if signal was received but status check failed
   docker compose logs backend | grep -A5 "SIGNAL RECEIVED"
   ```
   Look for "Transaction status check" log showing the actual status value

**Common Fixes:**

1. **Status code mismatch:** Ensure mockgatehub uses `TransactionStatusCompleted = 100`
2. **Webhook not sent:** Check mockgatehub webhook manager is running and configured
3. **Signal not delivered:** Verify payment_id mapping in gatehub_transactions table is correct
4. **Workflow ID mismatch:** Signal delivery uses `payment_pay_out_{payment_id}` format

### Debugging Workflow Issues

**1. Check if workflow is stuck waiting:**
```bash
# Get workflow status
docker exec local-temporal-1 temporal workflow describe \
  --workflow-id "payment_pay_out_<payment_id>" \
  --namespace default

# Look for:
# - Status: RUNNING (should be COMPLETED after payment)
# - PendingActivities: Empty or GetGatehubTransfer
# - PendingTimers: 20-minute timer indicates polling mode
```

**2. Trace webhook delivery path:**
```bash
# Step 1: Check if mockgatehub sent webhook
docker compose logs mockgatehub | grep "core.deposit.completed"

# Step 2: Check if backend received webhook
docker compose logs backend | grep "Webhook received"

# Step 3: Check if signal was sent
docker compose logs backend | grep "SignalGatehubTransferComplete"

# Step 4: Check if workflow received signal
docker compose logs backend | grep "SIGNAL RECEIVED"

# Step 5: Check what status was returned
docker compose logs backend | grep "Transaction status check"
```

**3. Verify gatehub_transactions mapping:**
```sql
-- Ensure external transaction ID maps to correct payment
SELECT payment_id, external_id 
FROM gatehub_transactions 
WHERE payment_id = '<payment_uuid>';

-- If mapping is missing, signal won't find the workflow
```

**4. Manual signal testing:**
```bash
# Send test signal to workflow (requires temporal CLI)
docker exec local-temporal-1 temporal workflow signal \
  --workflow-id "payment_pay_out_<payment_id>" \
  --name "payment_gatehub_signals" \
  --namespace default
  
# Check if workflow wakes up and queries transaction status
docker compose logs backend | tail -20
```

**5. Check transaction status directly:**
```bash
# Query mockgatehub transaction
curl http://localhost:8080/core/v1/transactions/<tx_uuid>

# Or via Redis
docker exec local-redis-1 redis-cli -n 2 GET "tx:<tx_uuid>" | jq '.'

# Verify status field is 100, not 1, 2, or other values
```

## GateHub Webhook and Signal Flow

### Overview

The payment workflow uses a hybrid approach to detect when external GateHub transactions complete:

1. **Webhook Notification (Preferred)**: GateHub/MockGatehub sends a webhook when transaction status changes
2. **Polling Fallback**: If webhook fails or is delayed, workflow polls every 20 minutes

This ensures payments complete reliably even if webhooks are missed or delayed.

### Webhook Flow

```
┌─────────────────────────────────────────────────────────────────────────┐
│                  GATEHUB WEBHOOK → TEMPORAL SIGNAL FLOW                 │
└─────────────────────────────────────────────────────────────────────────┘

GateHub/Mock          Backend Webhook Handler         Temporal Workflow
    │                          │                              │
    │  1. Transaction          │                              │
    │     completes            │                              │
    │     (status=100)         │                              │
    │                          │                              │
    │─ POST /webhooks/gatehub ─→                              │
    │   {                      │                              │
    │     event_type:          │                              │
    │       "core.deposit.     │                              │
    │        completed",       │                              │
    │     tx_uuid: "...",      │                              │
    │     deposit_type:        │                              │
    │       "hosted"           │                              │
    │   }                      │                              │
    │                          │                              │
    │                          │ 2. Route to handler          │
    │                          │    HandleUserDeposit()       │
    │                          │                              │
    │                          │ 3. Lookup payment_id         │
    │                          │    via gatehub_transactions  │
    │                          │    WHERE external_id=tx_uuid │
    │                          │                              │
    │                          │ 4. Construct workflow_id     │
    │                          │    = "payment_pay_out_       │
    │                          │       {payment_id}"          │
    │                          │                              │
    │                          │ 5. Send Temporal signal      │
    │                          │─ SignalWorkflow() ──────────→│
    │                          │    channel: payment_         │
    │                          │    gatehub_signals           │
    │                          │                              │
    │← HTTP 200 OK ─ ─ ─ ─ ─ ─│                              │
    │                          │                              │ 6. Selector wakes
    │                          │                              │    (signal received)
    │                          │                              │
    │                          │                              │ 7. Call activity:
    │                          │                              │    GetGatehubTransfer
    │                          │                              │
    │                          │← GetTransaction(tx_uuid) ── ─│
    │                          │                              │
    │─ GET /transactions/... ──→                              │
    │                          │                              │
    │← {status: 100, ...} ─ ─ ─│                              │
    │                          │                              │
    │                          │─ Return status=100 ─────────→│
    │                          │                              │
    │                          │                              │ 8. Check status
    │                          │                              │    if (status==100)
    │                          │                              │      break loop
    │                          │                              │
    │                          │                              │ 9. Continue workflow
    │                          │                              │    FinalizeBalance
    │                          │                              │    AssignBalance
    │                          │                              │    CreatePayoutTx
```

### Signal Channel Details

**Channel Name**: `payment_gatehub_signals`

**Signal Delivery**:
```go
// In backend/payments/ops/ops.go
func SignalGatehubTransferComplete(externalTransactionID string) error {
    // 1. Find payment_id from external_transaction_id
    row := db.QueryRow(
        "SELECT payment_id FROM gatehub_transactions WHERE external_id=$1",
        externalTransactionID,
    )
    
    // 2. Construct workflow ID
    workflowID := fmt.Sprintf("payment_pay_out_%s", paymentID)
    
    // 3. Send signal to Temporal
    client.SignalWorkflow(
        ctx,
        workflowID,
        "",  // empty runID means current run
        "payment_gatehub_signals",  // channel name
        nil,  // no payload needed
    )
}
```

**Signal Reception**:
```go
// In backend/payments/ops/workflows.go (PayoutWorkflow)
func PayoutWorkflow(ctx workflow.Context, req PaymentRequest) error {
    signalChannel := workflow.GetSignalChannel(ctx, "payment_gatehub_signals")
    
    for {
        selector := workflow.NewSelector(ctx)
        
        // Listen for signal OR timer
        selector.AddReceive(signalChannel, func(c workflow.ReceiveChannel, more bool) {
            c.Receive(ctx, nil)  // Signal received - wake up
        })
        
        timer := workflow.NewTimer(ctx, 20*time.Minute)
        selector.AddFuture(timer, func(f workflow.Future) {
            // Timer fired - wake up
        })
        
        selector.Select(ctx)  // Block until signal OR timer
        
        // Check transaction status
        var tx GatehubTransaction
        workflow.ExecuteActivity(ctx, GetGatehubTransfer, externalTxID).Get(ctx, &tx)
        
        if tx.Status == 100 {  // Completed
            break  // Exit polling loop
        }
        
        // Status != 100, loop back to wait again
    }
    
    // Continue with finalize, assign, create transaction...
}
```

### Transaction Status Codes

**Critical**: The workflow completion depends on receiving the correct status code.

| Status Code | Meaning | GateHub API | Workflow Action |
|------------|---------|-------------|-----------------|
| `1` | Pending/Processing | Standard | Continue polling |
| `100` | Completed | Standard | Break loop, complete payment |
| `3` | Failed | Standard | Mark payment failed |

**Common Issue**: If mockgatehub uses a non-standard status code (e.g., `2` instead of `100`), the workflow will receive the signal, check the status, see it's not `100`, and continue polling indefinitely.

**Validation**:
```bash
# Check transaction status in mockgatehub
docker exec local-redis-1 redis-cli -n 2 GET "tx:<uuid>" | jq '.status'

# Expected: 100 (not 1, 2, or other values)
```

### Polling Fallback Mechanism

If webhooks fail or are delayed, the workflow uses a 20-minute polling timer:

1. Timer expires after 20 minutes
2. Workflow calls `GetGatehubTransfer` activity
3. Checks transaction status
4. If status == 100, breaks loop
5. Otherwise, sets new 20-minute timer and waits again

This ensures eventual consistency even without webhooks.

**Trade-offs**:
- **Webhooks**: Fast (sub-second completion), but requires reliable network
- **Polling**: Slow (20-minute intervals), but guaranteed to work

## Temporal Workflow Monitoring

### View Workflow Status

Workflows are tracked by Temporal. To check payment workflow status:

1. Check Temporal UI (usually http://localhost:8233)
2. Search by payment ID as workflow ID
3. Look for child workflows:
   - `payment_payin_{payment_id}`
   - `payment_payout_{payment_id}`

### Workflow States
- **RUNNING**: Still processing
- **COMPLETED**: Success
- **FAILED**: Error occurred
- **TERMINATED**: Manually stopped

### Common Workflow Failures
- Insufficient balance (non-retryable)
- Receiver not ready
- External provider timeout (retried)

## Payment Type Routing

Different payment types follow different paths:

```
Payment Type        Provider Route           Balance Impact              Transaction Creation
─────────────────────────────────────────────────────────────────────────────────────────────
P2P                 GateHub/Xago/PTI        Reserve → Finalize → Assign    Local Workflow
Withdrawal          GateHub/Xago/PTI        Reserve → Finalize (no assign) Local Workflow
Deposit             GateHub/Xago/PTI        (no reserve) → Assign          Local Workflow
RafikiPeer2Peer     Rafiki webhook*         Reserve → Finalize → Assign    Rafiki Webhook**
WebMonetization     Rafiki webhook          Reserved via Rafiki             Rafiki Webhook**
Rafiki2External     Rafiki webhook*         Reserve → Finalize (no assign) Rafiki Webhook**

* Rafiki webhook creates payment record and may immediately complete or defer
** Receive transaction is created by Rafiki webhook handler (immediatePayment),
   NOT by the CreatePayoutTransaction activity in the payment workflow.
   This is crucial because receiver is an external wallet URL, not a local wallet.
```

### Receiver Identification Types

The `Receiver.Type` determines how transaction creation works:

```
Identity Type              Example                    Transaction Created By
───────────────────────────────────────────────────────────────────────────
WalletID                   [uuid]                     CreatePayoutTransaction (workflow)
WalletURL                  http://localhost/wallet    CreatePayoutTransaction (workflow)
ExternalWalletURL          https://other.ilp.link     Rafiki webhook handler ⚠️
Email/Twitter/Slack/etc    alice@example.com          CreatePayoutTransaction (after identity resolved)
```

⚠️ **Critical**: For `ExternalWalletURL`, `CreatePayoutTransaction` skips creation and returns early.
The receive transaction will be created asynchronously when the Rafiki webhook fires.

## Email Notifications

After payment completes:

1. **Sender receives:**
   - "Payment Sent" email with amount, recipient, date
   - Links to transaction details

2. **Receiver receives:**
   - "Payment Received" email with amount, sender, date
   - Links to transaction details

3. **On Failure:**
   - Sender receives "Payment Failed" email

## Performance Considerations

1. **Concurrent Payments**: PayIn and PayOut workflows run in parallel
2. **Provider Calls**: Each provider (GateHub, Xago, PTI) has 20-second timeouts
3. **Database Indexes**: Key queries use indexes on (wallet_id, state), (foreign_id)
4. **Temporal Retries**: Failed activities retry with exponential backoff

## Related Services

- **GateHub**: Manages balances, accounts, KYC
- **Rafiki**: Open payments protocol for ILP
- **Temporal**: Orchestrates multi-step workflows
- **PostgreSQL**: Stores all payment/transaction data
- **MockGatehub**: Local development mock (port 8080)

---

*Last updated: January 27, 2026*
*Updated with webhook/signal flow and transaction status code documentation*
