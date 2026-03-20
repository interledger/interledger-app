# Phase 7 — Fee Estimation and Currency Conversion

**Goal**: The two info endpoints used by the backend.

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

- [ ] Fee estimation with all fields succeeds
- [ ] Fee estimation requires `amount`
- [ ] Fee estimation without rail requires `currency` = USD
- [ ] Fee estimation with non-USD currency and no rail returns 400
- [ ] Fee direction defaults to `payout`
- [ ] Fee direction `funding` is accepted
- [ ] Fee is consistent across identical requests
- [ ] Fee reflects configured `INTERAC_FEE_FLAT`
- [ ] `netAmount` equals `amount` minus `totalFee`
- [ ] Currency conversion CAD to USD succeeds
- [ ] Conversion result reflects configured `CAD_TO_USD_RATE`
- [ ] Conversion requires `originCurrency`
- [ ] Conversion requires `amountInOriginCurrency`
- [ ] Conversion with zero amount returns zero USD
- [ ] Conversion is repeatable with same rate
- [ ] All previous feature scenarios still pass
- [ ] All code passes `golangci-lint run ./...`
- [ ] Unit test coverage ≥ 60% (unit tests use memory-backed store); E2E tests use redis-backed store
- [ ] `make test` runs successfully (linting + unit tests + e2e tests)
