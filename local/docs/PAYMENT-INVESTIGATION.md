# Payment Investigation: dabatla@gmail.com → alice@example.com

## Issue
Payment sent from dabatla@gmail.com to alice@example.com is not showing up on Alice's side.

## Investigation Findings

### 1. Database State

**Payment Record** (ID: `96c8cc12-3f0d-4928-9891-70bc4c07c1fa`):
```
public_id: 85erj6cvfzzm
state: 3 (COMPLETED)
sender_id: f32008e6-543c-4a3d-8229-d68f2ce8b899
receiver_id: https://local.ilp.link/alice
receiver_id_type: 3 (IdentityTypeWalletURL)
sender_amount: 2200 EUR
receiver_amount: 2200 EUR
sender_account: 1f46bdb0-6ac3-4f6d-a0f6-5dac4173b921 ✓
receiver_account: e2025323-28de-40eb-8821-1a148de96a4c ✓
send_transaction_id: 2ee2856d-7616-4ef1-9784-dc524eced88d ✓
receive_transaction_id: NULL ✗ ← PROBLEM!
type: 1 (Peer2Peer)
```

**Sender Transaction** (ID: `2ee2856d-7616-4ef1-9784-dc524eced88d`):
```
wallet_id: f32008e6-543c-4a3d-8229-d68f2ce8b899 (stephan)
type: sent
state: Pending ✗ ← Should be 'completed'
amount: 2200 EUR
Transfers: NONE ✗ ← No transfers created!
```

**Receiver Transaction**: Does NOT exist (receive_transaction_id is NULL)

### 2. User & Wallet Configuration

**Users in Kratos:**
- dabatla@gmail.com: `07a2d812-f214-44e0-9e2b-363071964de8` ✓
- alice@example.com: `768d6396-e981-4508-a7f9-72cd271d4bb5` ✓

**Wallets:**
- stephan (dabatla): `f32008e6-543c-4a3d-8229-d68f2ce8b899` ✓
- alice: `fcccd34d-60a9-46dc-9c39-e327bc99d2cc` ✓

**Wallet Addresses (Rafiki):**
- stephan: `https://local.ilp.link/stephan` ✓
- alice: `https://local.ilp.link/alice` ✓

**Linked Accounts (GateHub):**
- Sender account: `1f46bdb0-6ac3-4f6d-a0f6-5dac4173b921` (gatehub, EUR, balance) ✓
- Receiver account: `e2025323-28de-40eb-8821-1a148de96a4c` (gatehub, EUR, balance) ✓

### 3. Rafiki State

**Incoming Payments**: NONE ✗
**Outgoing Payments**: NONE ✗

This is the key issue: No Rafiki payments were created at all!

### 4. MockGatehub State

**Service Status**: Running and healthy ✓
**Configuration**: 
- Accessible via Traefik at `mockgatehub.interledger.test`
- Using Redis DB 2
- Webhooks configured to `http://backend:8080/webhooks/gatehub`
- Authentication disabled for local dev

**Users**: Present in Redis ✓

## Root Cause Analysis

The payment was created with:
- `receiver_id`: `https://local.ilp.link/alice` (local Rafiki wallet URL)
- `receiver_id_type`: `3` (WalletURL, not ExternalWalletURL)
- Both sender and receiver accounts are local GateHub accounts

This should trigger a **local peer-to-peer payment via Rafiki**, but the Rafiki payment flow never started:

1. ✓ Payment record created in wallet backend
2. ✓ Sender and receiver accounts identified (both GateHub EUR)
3. ✗ **Rafiki outgoing payment NOT created** ← CRITICAL FAILURE
4. ✗ Sender transaction has no transfers
5. ✗ Sender transaction stuck in "Pending" state
6. ✗ Receiver transaction never created

## Why Rafiki Wasn't Invoked

Looking at the payment type (1 = Peer2Peer) and receiver identity type (3 = WalletURL), the payment should flow through one of these paths:

### Path A: Direct GateHub P2P (if both local wallets with same provider)
- Reserve sender balance
- Finalize sender balance  
- Assign receiver balance
- Create both transactions

### Path B: Rafiki P2P (if using wallet URLs for ILP)
- Create Rafiki outgoing payment
- Rafiki processes the payment
- Rafiki webhook triggers receiver transaction creation

The payment has `receiver_id` as a wallet URL (`https://local.ilp.link/alice`), suggesting it should use **Path B**, but:
- No Rafiki outgoing payment was created
- The workflow appears to have not completed

## Hypothesis: Workflow Never Completed

The most likely issue is that the Temporal workflow that processes the payment never completed successfully. Evidence:

1. Sender transaction is in "Pending" state (not "Completed")
2. No transfers were created for the sender transaction
3. No receive transaction was created
4. Payment state shows as "Completed" but transactions don't match this

This suggests:
- Payment was created and marked as processing
- Workflow started but failed at an early stage
- Payment was incorrectly marked as completed despite workflow failure
- No transfers were actually executed

## Next Steps for User

1. **Check Temporal Workflow Logs**
   ```bash
   # Look for workflow execution for payment ID
   # Workflow ID format: payment_{payment_id}
   ```

2. **Check Backend Logs for Errors**
   ```bash
   docker logs $(docker ps --filter "name=local-wallet-backend" -q) 2>&1 | grep "96c8cc12"
   ```

3. **Verify Temporal is Running and Processing Workflows**
   ```bash
   docker ps | grep temporal
   # Check Temporal UI at http://localhost:8233
   ```

4. **Check if Payment Type Should Be RafikiPeer2Peer Instead of Peer2Peer**
   - The payment might be incorrectly typed
   - With receiver as wallet URL, should it be TypeRafikiPeer2Peer (5)?

5. **Check Payment Creation Logic**
   - How does the system decide between Peer2Peer and RafikiPeer2Peer?
   - When receiver is a wallet URL, what type should be used?

## Configuration Issues to Check

### Is Rafiki Integration Enabled?
Check if the wallet backend is configured to use Rafiki for wallet URL receivers.

### Are Wallet URLs Supposed to Use Rafiki?
The system has both:
- Local GateHub accounts for both users
- Rafiki wallet addresses for both users

Question: When should Rafiki be used vs direct GateHub transfer?

### Payment Type Assignment Logic
Need to review how payment type is determined when creating a payment to a wallet URL receiver.

---

**Conclusion**: The issue is NOT with the backend code (which works in production/sandbox). The issue is likely:
1. Temporal workflow not completing
2. Payment type incorrectly set to Peer2Peer instead of RafikiPeer2Peer
3. Or configuration mismatch between expected payment flow and actual setup

The receiver IS identified correctly as a local wallet with a valid GateHub account. The payment flow just never executed the actual money transfer.

## Latest Investigation (Jan 26, 2026 - 15:00 UTC)

### New Payment Test: a78553c7-8c10-40e2-9ea2-1a44f4b76263

**Observation:** Both PayIn and PayOut workflows stuck in RUNNING state with no pending activities.

**Root Cause Identified:**

The PayIn workflow is waiting for a GateHub webhook or 20-minute polling cycle:

1. **Transaction Created Successfully:**
   - External ID: `4784551e-c35b-49b9-a404-adc7de4578b5`
   - Created at: 2026-01-26 15:01:00
   - Type: 2 (hosted transfer)
   - Amount: 12.00 EUR
   - Status: 1 (pending/processing)

2. **Workflow Behavior:**
   - After `GatehubTransfer` activity creates the transaction
   - Workflow enters a loop waiting for transaction completion
   - Two exit conditions:
     - Receive webhook signal on channel `gatehubNotifyChanName`
     - 20-minute timer expires → poll via `GetGatehubTransfer` activity
   - Loop continues until transaction status = completed (status=2)

3. **MockGatehub Issue:**
   - Transaction exists in Redis with status=1
   - MockGatehub does NOT automatically complete hosted transfers
   - MockGatehub does NOT send `core.deposit.completed` webhook
   - Result: Workflow polls indefinitely every 20 minutes

4. **Temporal Evidence:**
   ```
   Last log: NewTimer, TimerID="37", Duration=1200 (20 minutes)
   Workflow history event 37: TimerStarted
   No pending activities (waiting on timer or signal)
   ```

**Impact:**
- Payment appears complete in database (state=3)
- Workflows never close
- Resources held indefinitely
- No error visible to user

**Fix Required:**
MockGatehub needs to either:
1. Auto-complete hosted transfers on creation
2. Implement webhook delivery for transaction completion
3. Add async worker to transition transactions to completed status

## Next Steps

1. ✅ Vault ID fix validated - payments now create GateHub transactions
2. ⚠️  New issue: Workflows stuck waiting for webhook/polling
3. TODO: Fix mockgatehub to complete hosted transfers automatically
4. TODO: Implement webhook delivery for `core.deposit.completed`
5. Test with fresh payment after mockgatehub fix
