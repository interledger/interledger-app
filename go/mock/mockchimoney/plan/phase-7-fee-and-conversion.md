# Phase 7 — Fee Estimation and Currency Conversion

**Goal**: The two info endpoints used by the backend.

**Status**: Completed (2026-03-20)

**Definition of Done**: This phase is complete when `make test` passes successfully.

## Deliverables

- `internal/handler/fee.go` — `POST /v0.2.4/info/fee-estimate`:
  - Validates `amount` required (number > 0)
  - When `rail` is absent, `currency` must be USD (per official docs)
  - Computes `totalFee = INTERAC_FEE_FLAT` (flat fee; currency/rail ignored beyond validation)
  - Returns `{ amount, currency, rail, direction, totalFee, netAmount }`
- `internal/handler/convert.go` — `GET /v0.2.4/info/convert/local-amount-to-usd`:
  - Validates `originCurrency` and `amountInOriginCurrency` query params required
  - Computes `amountInUSD = amountInOriginCurrency * CAD_TO_USD_RATE` (rate from config)
  - Returns `{ originCurrency, amountInOriginCurrency, amountInUSD, validUntil, expiresAt, expiresAtTimestamp }`

## Feature Files

- `features/fee_estimation.feature`
- `features/currency_conversion.feature`

## Dependencies

- Phases 0–6 must be complete and green

## Test-Driven Development Notes

1. **Red**: Run fee and conversion scenarios. Expect handler not found errors.
2. **Green**:
   - Implement handlers with basic math and validation.
   - Use config values for fee and exchange rate.
3. **Refactor**:
   - Config-driven fee/rate values should be clearly testable by changing env vars.
   - Extract calculation logic into testable utility functions.
   - Run all previous feature scenarios.

## Acceptance Criteria

- [x] Fee estimation with all fields succeeds
- [x] Fee estimation requires `amount`
- [x] Fee estimation without rail requires `currency` = USD
- [x] Fee estimation with non-USD currency and no rail returns 400
- [x] Fee direction defaults to `payout`
- [x] Fee direction `funding` is accepted
- [x] Fee is consistent across identical requests
- [x] Fee reflects configured `INTERAC_FEE_FLAT`
- [x] `netAmount` equals `amount` minus `totalFee`
- [x] Currency conversion CAD to USD succeeds
- [x] Conversion result reflects configured `CAD_TO_USD_RATE`
- [x] Conversion requires `originCurrency`
- [x] Conversion requires `amountInOriginCurrency`
- [x] Conversion with zero amount returns zero USD
- [x] Conversion is repeatable with same rate
- [x] All previous feature scenarios still pass
- [x] All code passes `golangci-lint run ./...`
- [x] Unit test coverage ≥ 60% (unit tests use memory-backed store); E2E tests use redis-backed store
- [x] `make test` runs successfully (linting + unit tests + e2e tests)

## Implementation Notes

- Added `internal/handler/fee.go` implementing `POST /v0.2.4/info/fee-estimate` with amount validation, USD rule when rail is omitted, default direction, and config-driven fee.
- Added `internal/handler/convert.go` implementing `GET /v0.2.4/info/convert/local-amount-to-usd` with required query params and config-driven exchange conversion.
- Added unit coverage in `internal/handler/info_test.go`.

## Phase 8 Handoff Notes

- KYC handlers can reuse webhook enqueueing and shared sub-account lookup/update logic already present in storage and handler helpers.
