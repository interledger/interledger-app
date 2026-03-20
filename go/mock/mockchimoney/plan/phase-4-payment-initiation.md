# Phase 4 — Payment Initiation and Verification

**Goal**: The deposit half of the payment flow: creating a payment link, serving the pay page, and verifying payment status.

**Status**: Completed (2026-03-20)

**Definition of Done**: This phase is complete when `make test` passes successfully.

## Deliverables

- `internal/storage/interface.go` extended: `CreatePayment`, `GetPaymentByIssueID`, `UpdatePaymentStatus`
- `internal/storage/memory.go` updated
- `internal/handler/payment.go`:
  - `POST /v0.2.4/payment/initiate` — validates `payerEmail` (required), `currency` ∈ {USD, NGN, CAD}, generates `issueID` as `{chiWalletID}_{uuid}`
  - `POST /v0.2.4/payment/verify` — looks up by `id` field, returns `Payment` object including `meta.processingFee`
  - `GET /pay/<issueID>` — reads embed of `web/pay.html`; 404 for unknown issueID
  - `POST /pay/<issueID>/confirm` — marks payment as `redeemed`; redirects to `redirect_url?issueID=...&status=success`
- `web/pay.html` — minimal HTML form with a single "Pay now" button that POSTs to `/pay/<issueID>/confirm`

## Feature Files

`features/deposits_and_webhooks.feature` (initiate, verify, and pay-page scenarios only; webhook scenarios in Phase 5)

## Dependencies

- Phases 0–3 must be complete and green

## Phase 3 Handoff Notes

- Reuse `internal/handler/wallet_validation.go` helpers in payment handlers:
   - `requireTrimmedField` for required request fields.
   - `ensureSubAccountExists` for validating `subAccount` before creating payment records.
- Keep validation error wording consistent with existing handlers (`"<field> is required"`) to match existing feature assertions.
- Route registration pattern is in `cmd/mockchimoney/main.go` inside the authenticated route group.
- Current baseline test gate is green (`make test`) with total unit coverage at `63.2%`; maintain this while adding payment-phase code.

## Test-Driven Development Notes

1. **Red**: Run feature scenarios for deposit initiation and verification. Expect handler not found errors.
2. **Green**: 
   - Implement storage methods for payment CRUD.
   - Implement handlers with basic validation and JSON responses.
   - Serve `web/pay.html` as embedded HTML.
   - Generate issueID in format `{chiWalletID}_{uuid}`.
3. **Refactor**:
   - `issueID` generation must be in a single utility function (reused in payouts in Phase 6).
   - `web/pay.html` should be embedded via `//go:embed` directive.
   - Create helpers for request parsing and response shaping.
   - Run all previous feature scenarios.

## Acceptance Criteria

- [x] Deposit initiation with all required fields succeeds
- [x] Deposit initiation without `payerEmail` returns 400
- [x] Deposit initiation with unsupported currency (not USD/NGN/CAD) returns 400
- [x] Deposit initiation with USD is accepted
- [x] Deposit initiation with NGN is accepted
- [x] Deposit initiation without `amount` returns 400
- [x] Deposit initiation for non-existent sub-account returns 400
- [x] Payment is verified as "pending" before pay page completion
- [x] Pay page is served as HTML for valid issueID
- [x] Pay page returns 404 for unknown issueID
- [x] Completing payment redirects to `redirect_url` with `issueID` and `status=success`
- [x] Payment is marked "redeemed" after pay page completion
- [x] Verify response includes fee metadata
- [x] All previous feature scenarios still pass
- [x] All code passes `golangci-lint run ./...`
- [x] Unit test coverage ≥ 50% (unit tests use memory-backed store)
- [x] `make test` runs successfully (linting + unit tests + e2e tests)

## Implementation Notes

- Added payment storage methods in `internal/storage/interface.go` and `internal/storage/memory.go`.
- Added `internal/handler/payment.go` with initiate, verify, pay-page, and confirm endpoints.
- Added embedded pay page template in `web/pay.html` via `web/templates.go`.
- Added issueID generation utility shared for payments and payouts (`internal/handler/common.go`).
- Added payment unit tests in `internal/handler/payment_test.go`.

## Phase 5 Handoff Notes

- `ConfirmPayPage` now updates payment status to `redeemed` and has webhook enqueue hooks ready.
- The webhook sender and queue integration should be validated against svix signature behavior and sequencing requirements.
