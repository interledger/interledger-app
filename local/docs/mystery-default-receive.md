# Mystery: `default_receive` for GateHub linked accounts

## Summary
Protea payment creation does **not** pass `receiverAccount`, but the backend can still set `receiver_account` during payment creation by **auto-pairing** sender/receiver balance accounts with matching currency. In sandbox, this auto-pairing is what makes GateHub payments work even though GateHub accounts have `default_receive=false`. The mystery is therefore **not** whether `default_receive` is set (it isn’t) but whether **currency-matched GateHub balance accounts exist at creation time**.

## Key observations (local)
- **Protea pay flow doesn’t provide `receiverAccount`.** It only passes the receiver identity (wallet URL).
  - [typescript/protea/app/routes/pay/route.tsx](../../typescript/protea/app/routes/pay/route.tsx#L151-L180)
  - [typescript/protea/app/routes/me_.$.tsx](../../typescript/protea/app/routes/me_.$.tsx#L252-L282)

- **Backend auto-pairs sender/receiver balance accounts when both account IDs are empty.**
  - [go/backend/payments/ops/ops.go](../../go/backend/payments/ops/ops.go#L338-L386)

- **Backend uses `GetDefaultReceive` when `ReceiverAccount` is empty.**
  - [go/backend/payments/ops/ops.go](../../go/backend/payments/ops/ops.go#L1040-L1055)

- **`GetDefaultReceive` fallback excludes GateHub.** Only PTI/Xago are eligible when `default_receive=false`.
  - [go/backend/linkedaccounts/ops/service.go](../../go/backend/linkedaccounts/ops/service.go#L314-L345)

- **GateHub linked accounts are created without `default_receive=true`.**
  - [go/backend/providers/gatehub/ops/activity.go](../../go/backend/providers/gatehub/ops/activity.go#L130-L165)

- **`SetDefaultReceive` rejects GateHub provider.**
  - [go/backend/linkedaccounts/ops/service.go](../../go/backend/linkedaccounts/ops/service.go#L618-L634)

- **Local data confirms zero GateHub defaults:**
  - `SELECT count(*) FROM linked_accounts WHERE provider='gatehub' AND default_receive=true;` → **0**

## Sandbox findings (Dec 2025 data)
- **`receiver_account` is populated on payments**, even though `default_receive=false` on GateHub accounts.
- **`receiver_account` values map to GateHub balance accounts** with `can_receive=true`, `state=Verified`, `receive_currency=EUR`, and `default_receive=false`.
- **All sampled GateHub receiver accounts are EUR** (matching sender currency), which enables auto-pairing during payment creation.

This indicates sandbox payments succeed because the **auto-pairing logic** selects matching GateHub balance accounts at creation time, not because `default_receive` is set.

## Why production/sandbox still works (revised)
1. **Auto-pairing at payment creation** picks sender/receiver GateHub balance accounts when currencies match.
2. **Explicit `receiver_account`** is not required if currency-matched balances exist.
3. **If currencies do not match**, the system falls back to `GetDefaultReceive` (which excludes GateHub), causing the local failure mode.

## Next checks
- Verify **sender/receiver GateHub balances exist for the same currency** in local (e.g., EUR↔EUR vs USD↔EUR).
- Inspect whether **local payments are created with mismatched currencies**, which would bypass auto-pairing and force `GetDefaultReceive`.
- Confirm whether any local flow **explicitly sets** `receiver_account`.

## Feature enablement mystery (CLI vs manual user creation)

### Observation
- **Manually created users** (e.g., dabatla@gmail.com in botanist admin portal): All features enabled
- **CLI-created users** (e.g., randomuser1769502238@example.com): Only `AccountEnabled=true`, all other features disabled

### Root cause: Missing KYC status
The `Features()` function in [go/backend/features/ops/ops.go](../../go/backend/features/ops/ops.go#L63-L68) gates all feature enablement on KYC status:

```go
if kycStatus != kyc.StatusLevel1 && kycStatus != kyc.StatusLevel2 {
    return features, nil  // Returns AccountEnabled=true, all other features false
}
```

**CLI user creation flow** ([local/scripts/internal/user/client.go](../scripts/internal/user/client.go)):
- ✅ Creates Kratos identity (with valid `tel:+1949xxxxxxx` phone format)
- ✅ Creates wallet and GateHub linked account (EUR balance, `can_receive=true`, `default_receive=false`, `state=Verified`)
- ❌ **Does NOT call `SetKYCStatus`** — `setupKYC()` and `activateUser()` are no-ops (lines 337-368)

**Manual user creation** (via admin/botanist):
- Explicitly sets KYC status to Level1 or Level2 during user activation
- Triggers country-based feature defaults (US: SendEnabled/ReceiveEnabled/BanksEnabled, EU: ManageWalletCardsEnabled, etc.)

### Impact
- CLI-created users have **functional GateHub linked accounts** but **disabled wallet features**
- Payments may fail or behave unexpectedly if sender/receiver features are required for transaction flow
- Auto-pairing logic still works (currency matching), but UI/API may block operations based on feature flags

### Fix required
Update [local/scripts/internal/user/client.go](../scripts/internal/user/client.go) `activateUser()` function to call:
- `SetKYCStatus(kyc.StatusLevel1)` after GateHub user creation
- This will trigger feature computation with country-based defaults

### Database verification
```sql
-- Check KYC status (missing for CLI users)
SELECT w.id, w.email, k.status 
FROM wallets w 
LEFT JOIN kyc k ON w.id = k.wallet_id 
WHERE w.email = 'randomuser1769502238@example.com';
-- Result: kyc.status is NULL

-- Check features (all disabled except AccountEnabled)
SELECT * FROM wallet_features WHERE wallet_id = (
    SELECT id FROM wallets WHERE email = 'randomuser1769502238@example.com'
);
-- Result: No row exists (features computed from KYC status at runtime)
```

## Potential fixes (code)
Pick **one** of the following to reconcile behavior for mismatched currencies / missing balances:
1. **Set `DefaultReceive=true`** on GateHub linked account creation.
2. Allow **GateHub in `SetDefaultReceive`** eligibility.
3. Add **GateHub to `GetDefaultReceive` fallback**.

**REQUIRED FIX for CLI user creation:**
- Add `SetKYCStatus(kyc.StatusLevel1)` call in `activateUser()` to enable wallet features for CLI-created users.

These are not implemented yet—this document is a tracking note.