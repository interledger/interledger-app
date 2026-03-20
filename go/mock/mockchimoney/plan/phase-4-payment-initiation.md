# Phase 4 — Payment Initiation and Verification

**Goal**: The deposit half of the payment flow: creating a payment link, serving the pay page, and verifying payment status.

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

- [ ] Deposit initiation with all required fields succeeds
- [ ] Deposit initiation without `payerEmail` returns 400
- [ ] Deposit initiation with unsupported currency (not USD/NGN/CAD) returns 400
- [ ] Deposit initiation with USD is accepted
- [ ] Deposit initiation with NGN is accepted
- [ ] Deposit initiation without `amount` returns 400
- [ ] Deposit initiation for non-existent sub-account returns 400
- [ ] Payment is verified as "pending" before pay page completion
- [ ] Pay page is served as HTML for valid issueID
- [ ] Pay page returns 404 for unknown issueID
- [ ] Completing payment redirects to `redirect_url` with `issueID` and `status=success`
- [ ] Payment is marked "redeemed" after pay page completion
- [ ] Verify response includes fee metadata
- [ ] All previous feature scenarios still pass
- [ ] All code passes `golangci-lint run ./...`
- [ ] Unit test coverage ≥ 50% (unit tests use memory-backed store)
- [ ] `make test` runs successfully (linting + unit tests + e2e tests)
