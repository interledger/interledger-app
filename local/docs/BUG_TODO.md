# Bug TODO: GateHub PayIn Fails Without Vault ID

## Summary
- **Symptom:** Payment PayIn workflow stuck on `GatehubTransfer` activity with error: `either currency or vault_uuid is required` from mockgatehub `/core/v1/transactions`.
- **Root cause:** Backend calls `external.CreateTransactionRequest` with `VaultID` from `GATEHUB_PAYWISER_EURO_VAULT_ID`. In local env this env var was unset, so payload omitted `vault_uuid`, triggering GateHub validation failure.
- **Evidence:** Backend logs showed `GATEHUB_PAYWISER_EURO_VAULT_ID is set to is not set` and repeated activity failures. Temporal PayIn workflow retried >50 times until vault ID was added.

## Impact
- PayIn never completes; PayOut never starts fully; `receive_transaction_id` remains NULL; sender transaction stays pending.
- Affects any EUR PayIn in local env (and potentially other currencies if their vault env vars are missing).

## Repro (local)
1) Start local stack without `GATEHUB_PAYWISER_EURO_VAULT_ID` set for backend.
2) Send payment (e.g., dabatla → alice).
3) Temporal shows `GatehubTransfer` failures with `either currency or vault_uuid is required`.

## Current Mitigation
- Set `GATEHUB_PAYWISER_EURO_VAULT_ID=a09a0a2c-1a3a-44c5-a1b9-603a6eea9341` in `local/wallet.yaml` and rebuild backend. After setting, PayIn retries succeed and GateHub transaction is created.

## Suggested Backend Fix (do not implement here)
- Make vault ID mandatory at startup:
  - Validate `GATEHUB_PAYWISER_EURO_VAULT_ID` (and any other currency vault vars) on init; fail fast with a clear error if missing.
- Add payload fallback to avoid hard fail when env is absent:
  - If vault ID env is missing, derive `vault_uuid` from currency via mockgatehub `SandboxVaultIDs` map (EUR → `a09a0a2c-1a3a-44c5-a1b9-603a6eea9341`), or include `currency` field explicitly when calling `CreateTransaction` so GateHub accepts the request without vault UUID.
- Improve logging/monitoring:
  - Log the effective vault ID at startup; add a distinct non-retryable Temporal error when vault ID is empty to avoid infinite retries.

## Follow-ups
- Confirm all currency vault env vars are populated in local/prod configs.
- Re-run the affected payment to ensure `receive_transaction_id` is populated and both workflows close.
