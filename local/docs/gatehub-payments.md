# Gatehub Payments Workflow

## Overview

This document describes the payment workflow in the Interledger App Wallet, specifically focusing on peer-to-peer (P2P) payments using Gatehub as a provider. It covers the complete journey of a payment from sender to receiver, including database interactions and how to verify payment status.

### Recent Fix (Jan 2026)

**Issue**: Payments with external wallet receivers (Rafiki) were not creating receive transactions, leaving money on sender's side without reaching receiver.

**Root Cause**: `CreatePayoutTransaction` tried to look up external wallet URLs as local wallets, causing failures.

**Solution**: Skip transaction creation for external wallet URLs. Receive transactions are created by Rafiki webhook handlers instead.

**Code Change**: [transactionactivities.go#L214-L216](transactionactivities.go) - Added check to skip transaction creation for `IdentityTypeExternalWalletURL`.

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
   │    ⚠️  Waits for webhook OR polls every 20 min          │                │
   │    - Webhook: core.deposit.completed                    │                │
   │    - Poll: GetGatehubTransfer activity                  │                │
   │    - Loop until status = 'completed'                    │                │
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

### Issue 4: PayIn Workflow Stuck Waiting for GateHub Webhook

**Symptoms:**
- Payment shows state=3 (COMPLETED) but workflows still RUNNING
- PayIn and PayOut workflows have no pending activities but don't close
- Temporal shows a 20-minute timer active in PayIn workflow
- GateHub transaction exists in mockgatehub with status=1 (pending)

**Root Cause:**
After creating a hosted transfer (type=2) in mockgatehub, the PayIn workflow waits for:
1. A webhook notification (`core.deposit.completed`) from mockgatehub, OR
2. 20-minute polling interval to check transaction status via `GetGatehubTransfer` activity

Mockgatehub currently:
- Creates the transaction successfully with status=1 (pending/processing)
- Does NOT automatically update hosted transfers to status=2 (completed)
- Does NOT send the `core.deposit.completed` webhook

This leaves the workflow in an infinite wait state, polling every 20 minutes indefinitely.

**How to Check:**
```sql
-- Check gatehub_transactions table for the external transaction ID
SELECT * FROM gatehub_transactions WHERE payment_id = 'payment-uuid';

-- Check transaction status in mockgatehub Redis (DB 2)
docker exec local-redis-1 redis-cli -n 2 GET "tx:<external_id>"
-- Look for "status":1 (pending) instead of "status":2 (completed)
```

**Temporal Evidence:**
```bash
# Check PayIn workflow - will show active timer
docker exec local-temporal-1 temporal workflow describe \
  --workflow-id "payment_pay_in_<payment_id>" \
  --namespace default --address temporal:7233

# Check workflow history - last event will be TimerStarted
docker exec local-temporal-1 temporal workflow show \
  --workflow-id "payment_pay_in_<payment_id>" \
  --namespace default --address temporal:7233 | tail -5
```

**Workaround:**
Manually update the transaction status in mockgatehub Redis:
```bash
# Get current transaction
docker exec local-redis-1 redis-cli -n 2 GET "tx:<external_id>"

# Update status from 1 to 2 (completed)
# Then trigger webhook manually or wait for next 20-minute poll
```

**Permanent Fix Needed in MockGatehub:**
1. Automatically set hosted transfers (type=2) to status=2 (completed) on creation
2. Send `core.deposit.completed` webhook immediately after transaction creation
3. Or implement an async worker that completes pending transactions after a short delay

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

*Last updated: January 2026*
