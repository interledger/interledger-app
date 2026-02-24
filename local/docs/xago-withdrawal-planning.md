# Xago Withdrawal Planning

## Scope

This document describes how the Xago withdrawal flow is intended to work across the Protea UI, backend workflows, MockXago, and the official Xago API documentation. It focuses on the South Africa (ZAR) balance-to-bank withdrawal path.

## Protea UI Flow (What the user does)

- The withdraw page is driven by [typescript/protea/app/routes/withdraw.tsx](typescript/protea/app/routes/withdraw.tsx).
- The loader calls `getOnOffRampProvider`. If the provider is `gatehub`, it loads the Gatehub iframe. Otherwise it loads the generic withdrawal UI (used for Xago/Interledger or PTI).
- For non-Gatehub providers, the UI pulls balances (`getBalancesForTransfer`) and eligible withdraw destinations (`getLinkedAccountsForWithdraw`), then renders an amount + bank selection flow.
- On submit, the UI posts to `grpc.withdrawBalance` with:
  - `fromLinkedAccount` = balance account
  - `toLinkedAccount` = bank account
  - `amount` in cents
  - currency code set as `ZAR` unless provider is `pti` (then `USD`)
- The action redirects to the confirm screen in [typescript/protea/app/routes/withdraw_.$paymentId.tsx](typescript/protea/app/routes/withdraw_.$paymentId.tsx) which calls `grpc.updatePayment` and `grpc.confirmPayment` to finalize the payment.

**Implication for Xago**: Xago uses the generic withdrawal flow (provider not `gatehub` and not `pti`). The UI sets currency code to ZAR and expects a valid Xago balance account plus at least one Xago bank account.

## Backend Flow (Payment + Xago Integration)

### 1) Create withdrawal payment

- The UI calls `WithdrawBalance` in [go/backend/grpc/withdraw.go](go/backend/grpc/withdraw.go).
- The handler verifies the linked accounts belong to the wallet and creates a payment with:
  - `Type = withdrawal`
  - Sender and receiver both set to the wallet
  - Sender account = balance account
  - Receiver account = bank account
- Validation for withdrawals happens in payments ops (only Xago, Chimoney, PTI supported): [go/backend/payments/ops/ops.go](go/backend/payments/ops/ops.go).

### 2) Reserve balance

- The payment workflow reserves balance for withdrawals via the provider-specific reserve logic. For Xago this is `Xago.ReserveBalance`, which uses Pacioli transfers: [go/backend/payments/ops/activities.go](go/backend/payments/ops/activities.go).

### 3) Payout phase (Xago)

- The payout workflow chooses provider-specific logic for the receiver account. If the receiver account is Xago bank, it uses `xagoPayOut`: [go/backend/payments/ops/workflows.go](go/backend/payments/ops/workflows.go).
- `xagoPayOut` does:
  1. `WithdrawFromXagoBalance` → `Xago.CreateTransaction` with the beneficiary ID (provider ID of the linked bank account).
  2. Poll `CheckXagoWithdrawalComplete` until Xago reports completion.
  3. Finalize the reserve and assign the balance update in Pacioli.
- The Xago activity uses:
  - `CreateTransaction` in [go/backend/providers/xago/ops/ops.go](go/backend/providers/xago/ops/ops.go)
  - `CheckXagoWithdrawalComplete` in [go/backend/payments/ops/activities.go](go/backend/payments/ops/activities.go)

### 4) Xago external API usage

- The external client calls:
  - `POST /v1/transfers` for creating a transfer (withdrawal): [go/backend/providers/xago/external/client.go](go/backend/providers/xago/external/client.go)
  - `GET /v1/transactions?transactionId=...` for withdrawal status: [go/backend/providers/xago/external/client.go](go/backend/providers/xago/external/client.go)
- The request payload supports idempotency keys and transaction type fields, but the client currently only sends `amount`, `currencyCode`, `beneficiaryId`, and `reference` (it does not set `idempotencyKey` or `transactionType`).

## Xago Bank Account (Beneficiary) Setup

- Xago withdrawals require a bank beneficiary (linked account type `bank_account`).
- Beneficiaries are created in the Xago workflow in [go/backend/providers/xago/ops/workflows.go](go/backend/providers/xago/ops/workflows.go):
  - `CreateBeneficiaryWorkflow` calls external `AddBeneficiary` and saves the record.
  - A linked account is created with `ProviderID = beneficiary ID` and `Type = bank_account`.
- The UI calls `AddXagoBankAccount` via gRPC to create the beneficiary: [go/backend/grpc/xago.go](go/backend/grpc/xago.go).

## MockXago Behavior

MockXago simulates the transfer and beneficiary endpoints that the backend expects.

- Transfer creation is handled by `POST /v1/transfers` in [go/mockxago/internal/handler/transactions.go](go/mockxago/internal/handler/transactions.go):
  - Validates `beneficiaryId`, checks balance, deducts funds, saves a transaction, and auto-completes after ~2.5s.
  - Status transitions from `pending` to `completed`.
- Withdrawal status lookup is handled by:
  - `GET /v1/transfers/{id}` or `GET /v1/transactions?transactionId=...` in [go/mockxago/internal/handler/transactions.go](go/mockxago/internal/handler/transactions.go).
- Beneficiary creation uses `POST /v1/accounts/{accountId}/beneficiaries` or the global `POST /v1/beneficiaries` alias in [go/mockxago/internal/handler/beneficiary.go](go/mockxago/internal/handler/beneficiary.go). Beneficiaries auto-approve after ~3s.

### MockXago vs Backend Expectations

- Backend waits for status `Success` (case-insensitive) in `CheckXagoWithdrawalComplete`.
- MockXago returns `completed`, so the backend will not mark withdrawals complete and will keep polling.

This is likely why the withdrawal E2E scenarios for Xago are tagged `@skip` in [local/e2e-playwright/features/004-withdrawal.feature](local/e2e-playwright/features/004-withdrawal.feature).

## Official Xago Documentation Notes

The official Xago Identity API docs (Postman) are published at:
- https://documenter.getpostman.com/view/49463771/2sB3QRo7pf

Key relevant items observed:

- `POST /v1/login` creates an access token (JWT). The response shape includes `tokenType` and `tokenValue`.
- Beneficiary listing is documented as `GET /v1/beneficiaries?limit=&page=`.

The documentation page is large and not all transfer endpoints were visible in the extracted content, but the internal mock and backend code assume:

- `POST /v1/transfers` for creating a withdrawal
- `GET /v1/transactions?transactionId=...` for checking transfer status
- `POST /v1/beneficiaries` (global) and `POST /v1/accounts/{accountId}/beneficiaries` (account-scoped)

These assumptions align with what MockXago implements and what the backend client calls.

## Current Gaps and Open Questions

1) **Status mismatch**
   - Backend expects `Success` but MockXago reports `completed`.
   - Result: Xago withdrawal polling never completes in local tests.

2) **Idempotency support**
   - Xago transfer types support `idempotencyKey`, but the external client does not pass one.
   - The payment workflow generates a transaction ID but it is not mapped into the request body.

3) **UI provider routing**
   - The UI uses the generic `withdrawBalance` path and sets currency to ZAR unless provider is `pti`.
   - There is a dedicated `WithdrawXagoBalance` endpoint in [go/backend/grpc/xago.go](go/backend/grpc/xago.go) that is not used by the UI.

4) **Official docs coverage**
   - The official Postman docs are large, and transfer endpoints were not visible in the scraped content. We should confirm the exact fields required for `POST /v1/transfers` and the status values returned by `GET /v1/transactions`.

## Suggested Next Steps

1) Confirm the official Xago transfer and transaction status semantics (especially status values).
2) Align MockXago status values with backend expectations (or relax backend status matching).
3) Decide whether the UI should call `WithdrawXagoBalance` instead of the generic `WithdrawBalance`.
4) Add idempotency support to the external client using the payment transaction ID.
