# GateHub Cross-Currency Payments (MockGatehub)

## Summary

Cross-currency payments are **not currently supported** end-to-end. The Interledger App blocks cross-currency payments at creation time, and the GateHub provider client only sends a single `vault_uuid` with no source/target currency pair. MockGatehub mirrors that single-currency flow.

## What the GateHub Docs Say

Based on the official GateHub documentation (linked in README), the Exchange flow is documented primarily as an **iframe product** for managed users rather than a direct transaction-creation API.

- GateHub Exchange is integrated via an iframe and emits an `ExchangeCompleted` message event to the parent app.
  - https://docs.gatehub.net/api-documentation/c3OPAp5dM191CDAdwyYS/gatehub-products/gatehub-exchange
- The Transactions reference lists **Exchange** as a transaction type, but the accessible pages do not show a create-exchange request schema or required FX fields.
  - https://docs.gatehub.net/api-documentation/c3OPAp5dM191CDAdwyYS/api-reference/api-reference/transactions

**Implication:** The official docs reinforce that FX is handled by the Exchange product (iframe), not by the existing `POST /core/v1/transactions` contract we emulate today.

## What the Codebases Say Today

### Interledger App (source of truth for wallet behavior)

- The payment pipeline explicitly rejects cross-currency, even when currencies differ. The FX helpers always return an error for cross-currency and for most non-USD cases.
  - [applyFXCreate](interledger-app/go/backend/payments/ops/ops.go#L625) returns `cross currency not supported`.
  - [applyFXUpdate](interledger-app/go/backend/payments/ops/ops.go#L1122) returns `cross currency currently not supported`.
- The GateHub external client only sends **one vault UUID** in `CreateTransactionRequest` and does not include a source currency, target currency, FX rate, or fee fields.
  - [CreateTransactionRequest](interledger-app/go/backend/providers/gatehub/external/types.go#L139)
- A constant for `TransactionTypeExchange = 3` exists, but there are no exchange endpoints or calls wired in.
  - [TransactionTypeExchange](interledger-app/go/backend/providers/gatehub/external/types.go#L17)

### MockGatehub (provider mock)

- `POST /core/v1/transactions` infers or validates a **single** currency using one vault UUID and updates a **single** currency balance.
  - [CreateTransaction handler](mockgatehub/internal/handler/core.go#L202)
  - [Vault inference from currency](mockgatehub/internal/handler/core.go#L250)
- Exchange rates are available via `/rates/current`, but they are **not used** in any transaction logic.
  - [GetCurrentRates](mockgatehub/internal/handler/rates.go#L12)
  - [Sandbox rates](mockgatehub/internal/consts/consts.go#L41)

**Conclusion:** Cross-currency payments do not work today and are not invoked by the app. MockGatehub is behaving correctly relative to current upstream behavior.

## Why Cross-Currency Is Ambiguous Right Now

The GateHub client only sends:
- `amount` (single numeric value)
- `vault_uuid` (single currency vault)
- `sending_address` / `receiving_address`
- `type`

There is **no way** to infer both the source currency and destination currency from this request alone. Without explicit source and target currency or vaults, MockGatehub cannot do deterministic FX conversion.

## Proposed MockGatehub Behavior When We Enable Cross-Currency

This is a forward-looking design that keeps compatibility with the existing single-currency flow and aligns with how GateHub exchange-style transactions are typically modeled.

### 1) Request Shape (explicit source + target)

Add optional fields to `CreateTransactionRequest` (accept without breaking existing callers):

- `sending_vault_uuid`
- `receiving_vault_uuid`
- `sending_amount` (string or float)
- `receiving_amount` (string or float)
- `fx_rate` (optional override)
- `fx_fee` or `fx_fee_percentage` (optional)

If `type == TransactionTypeExchange (3)`, require both vault UUIDs and one of:
- `sending_amount` (compute receiving amount), or
- `receiving_amount` (compute sending amount)

If `type != 3`, keep current behavior using `vault_uuid` only.

### 2) FX Rate Resolution

Use `SandboxRates` as the single source of truth:
- Rates are defined vs USD. To convert `A -> B`:
  - $rate_{A \to B} = rate_B / rate_A$

If a client supplies `fx_rate`, validate against the computed rate within a tolerance (or accept as authoritative for flexibility in tests).

### 3) Balance Updates

For exchange-style transactions:
- Debit the sender user balance in source currency.
- Credit the receiver user balance (or same user, depending on address) in destination currency.
- Store both amounts in the transaction record.

### 4) Transaction Representation

Extend `models.Transaction` to capture FX metadata:
- `SendingAmount`, `ReceivingAmount`
- `SendingCurrency`, `ReceivingCurrency`
- `FXRate`, `FXFee`, `FXFeePercentage`

This keeps `amount` and `currency` backwards compatible but makes FX auditable.

### 5) Webhooks and Status Codes

Keep the existing webhook event type (`core.deposit.completed`) for now to avoid breaking the wallet workflow, but include FX fields in the payload for new consumers. Ensure status reaches `100` as usual.

### 6) Test Updates

If/when enabled, update:
- Unit tests in MockGatehub for FX math and balance updates.
- Integration tests in mockgatehub `testenv/` to cover an exchange transaction type.
- Interledger App tests **only after** the backend begins to accept cross-currency requests.

## Recommendation

**Do not implement cross-currency in MockGatehub yet.** The Interledger App currently blocks it and the GateHub client does not send enough data to perform FX safely. The right sequence is:

1. Update Interledger App to accept cross-currency requests and send explicit source/target vaults or currencies.
2. Add exchange endpoints or request fields to the GateHub client layer.
3. Implement the FX flow in MockGatehub as outlined above.

Until then, MockGatehub’s current single-currency behavior is correct and prevents false confidence about FX support.

## Open Questions for Official GateHub Documentation

To finish the design, we still need to confirm (from GateHub docs):
- The exact request/response schema for exchange transactions (if `type=3` is used).
- Whether GateHub expects `vault_uuid` for source, destination, or both.
- Whether FX fees are returned in `fee`, `total_amount`, or separate FX-specific fields.
- The webhook event (type and payload) used for exchange completion.

These details should be verified against the official GateHub API reference before implementing.
