# Phase 11 (Optional) — BDD Test Suite

**Goal**: Automated Godog BDD tests that run against a real MockChimoney instance in Docker, suitable for CI.

**Definition of Done**: This phase is complete when `make test` passes successfully.

## Deliverables

- `testenv/` directory with the following structure:
  ```
  testenv/
  ├── godog_test.go
  ├── docker-compose.yml
  ├── http_client.go
  ├── test_context.go
  ├── helpers.go
  ├── wallet_steps.go
  ├── payment_steps.go
  ├── withdrawal_steps.go
  ├── kyc_steps.go
  └── webhook_server.go
  ```
- All feature scenarios that are not tagged `@wip` should pass
- `AGENTS.md` documenting how to run the test suite
- CI job or Makefile target: `make test-e2e`

## Feature Files

All feature files from Phases 0–10:
- `features/service_health.feature`
- `features/authentication.feature`
- `features/wallet_management.feature`
- `features/deposits_and_webhooks.feature`
- `features/withdrawals_and_webhooks.feature`
- `features/fee_estimation.feature`
- `features/currency_conversion.feature`
- `features/kyc.feature`
- `features/webhook_signing.feature`

## Dependencies

- Phases 0–10 must be complete and green
- Godog BDD framework
- Test HTTP client and helpers

## Test-Driven Development Notes

1. **Red**: Run the full feature suite with `go test -v ./testenv`. Expect many scenarios to fail.
2. **Green**:
   - Implement step definitions for all scenarios.
   - Implement helper functions to create wallets, initiate payments, etc.
   - Implement a webhook receiver that the mock can send to.
3. **Refactor**:
   - Deduplicate step definitions (e.g., "a wallet exists" should reuse wallet creation logic).
   - Ensure test fixtures (users, wallets) are cleaned up between scenarios.
   - Consider using a shared context object to track test state.
   - Run all scenarios to ensure coverage.

## Acceptance Criteria

- [ ] All non-`@wip` scenarios in all feature files pass
- [ ] `@wip` scenarios are skipped (not counted as failures)
- [ ] HTTP client handles JSON request/response marshalling
- [ ] Webhook receiver listens on a test port and captures payloads
- [ ] Signature verification in tests matches backend's verification logic
- [ ] Test context isolates scenarios (no state leakage between tests)
- [ ] All code passes `golangci-lint run ./...`
- [ ] CI job (`make test-e2e`) runs the full suite in Docker
- [ ] Unit test coverage ≥ 75% (unit tests use memory-backed store); E2E tests use redis-backed store
- [ ] `make test` runs successfully (linting + unit tests + e2e tests)

## AGENTS.md Structure

The `AGENTS.md` file should document:
- How to run individual feature files: `go test -v ./testenv -args -features features/service_health.feature`
- How to run a specific scenario by tag: `go test -v ./testenv -args -tags @payment_workflow`
- How to run with debug output: `go test -v ./testenv -args -debug -concurrency=1`
- Configuration environment variables
- Expected output format and examples
- Troubleshooting tips (e.g., "if webhook is not received, check WEBHOOK_URL env var")

## Notes

- This phase can be worked on in parallel with Phase 10 once the service is functionally complete after Phase 9.
- Consider adding a `@slow` tag to scenarios that involve waiting for webhooks, so the test suite can be split into fast unit tests and slow integration tests.
- The webhook receiver should capture signatures for validation testing (important for `webhook_signing.feature`).
