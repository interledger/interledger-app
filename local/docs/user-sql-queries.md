# User Status Query Recipes

Quick SQL recipes to diagnose user activation and Gatehub status.

## Find user by email and check activation status

```sql
-- Get wallet ID and KYC status for email
SELECT 
  s.email,
  s.user_id,
  s.first_name,
  s.last_name,
  w.id as wallet_id,
  w.country,
  wks.status as kyc_status,
  wks.created_at as kyc_set_at
FROM signups s
LEFT JOIN user_wallets uw ON s.user_id = uw.user_id
LEFT JOIN wallets w ON uw.wallet_id = w.id
LEFT JOIN wallet_kyc_status wks ON w.id = wks.wallet_id
WHERE s.email = 'user@example.com';
```

## Check Gatehub mapping and linked accounts for a wallet

```sql
-- Get Gatehub user ID and linked accounts
SELECT 
  w.id as wallet_id,
  w.name,
  w.country,
  gu.external_id as gatehub_user_id,
  gu.external_customer_id,
  gu.external_account_id,
  la.id as linked_account_id,
  la.provider,
  la.type,
  la.send_currency
FROM wallets w
LEFT JOIN gatehub_users gu ON w.id = gu.wallet_id
LEFT JOIN linked_accounts la ON w.id = la.wallet_id AND la.provider = 'gatehub'
WHERE w.id = 'wallet-uuid-here';
```

## List all wallets missing Gatehub mapping

```sql
-- Find wallets that clicked Activate but workflow failed
SELECT 
  w.id,
  w.name,
  w.country,
  gu.wallet_id,
  wks.status
FROM wallets w
LEFT JOIN gatehub_users gu ON w.id = gu.wallet_id
LEFT JOIN wallet_kyc_status wks ON w.id = wks.wallet_id
WHERE gu.wallet_id IS NULL
ORDER BY w.created_at DESC
LIMIT 10;
```

## Check Gatehub user creation attempts (Temporal)

From UI or CLI:
```bash
# View workflow execution history
temporal workflow describe \
  --workflow-id "gatehub_create_user_d24b610d-aae0-4324-a15b-543cb16b188d"

# In Temporal UI, search task queue "backend" for workflow ID containing wallet UUID
```

## One-liner: Full user diagnostics

```sql
-- Copy-paste this with your email to see everything
SELECT 
  s.email, w.id, w.country, gu.external_id, wks.status, 
  la.id, la.provider, la.send_currency
FROM signups s
LEFT JOIN user_wallets uw ON s.user_id = uw.user_id
LEFT JOIN wallets w ON uw.wallet_id = w.id
LEFT JOIN gatehub_users gu ON w.id = gu.wallet_id
LEFT JOIN wallet_kyc_status wks ON w.id = wks.wallet_id
LEFT JOIN linked_accounts la ON w.id = la.wallet_id AND la.provider = 'gatehub'
WHERE s.email = 'user@example.com';
```

## Connection string (local)

```bash
docker compose exec postgres psql -U postgres backend -c "YOUR_QUERY_HERE"
```
