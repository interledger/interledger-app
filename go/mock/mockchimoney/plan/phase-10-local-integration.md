# Phase 10 — Local Environment Integration

**Goal**: MockChimoney is part of the standard `local/` Docker Compose environment and the backend routes all chimoney API calls to it.

## Deliverables

- `local/mockchimoney.yaml` — Docker Compose service definition (see main plan.md section "Required Local Environment Changes")
- `local/docker-compose.yaml` — add `- mockchimoney.yaml` to `include:` list
- `local/wallet.yaml` — add the following env vars to the `backend` service:
  ```yaml
  - CHIMONEY_API_BASE_URL=${BACKEND_CHIMONEY_API_BASE_URL:-http://mockchimoney:8080/v0.2.4}
  - CHIMONEY_KYC_BASE_URL=${BACKEND_CHIMONEY_KYC_BASE_URL:-https://mockchimoney.interledger.test}
  - CHIMONEY_TOKEN=${BACKEND_CHIMONEY_TOKEN:-local-test-api-key}
  - CHIMONEY_WEBHOOK_SECRET=${BACKEND_CHIMONEY_WEBHOOK_SECRET:-local_bG9jYWwtdGVzdC13ZWJob29rLXNlY3JldA==}
  ```
- Backend changes: Add `CHIMONEY_API_BASE_URL` env var support in `go/backend/providers/chimoney/external/client.go` (see main plan.md section "Required Backend Changes")
- Update `local/example.env` with the new variables and their local defaults
- Smoke test: run `make all` in `local/`, create an Interledger wallet user, complete KYC, make a deposit, make a withdrawal — confirm all webhooks fire and Temporal workflows complete

## Feature Files

All previous feature files should pass as part of the local environment.

## Dependencies

- Phases 0–9 must be complete and green
- MockChimoney must be built and available as a Docker image
- Backend code must support the new env vars

## Test-Driven Development Notes

1. **Red**: Try to run `make all` in `local/`. Expect missing env var warnings or connection errors.
2. **Green**:
   - Create `mockchimoney.yaml` with proper Traefik/health check configuration.
   - Add the include to `docker-compose.yaml`.
   - Update `wallet.yaml` with the new env vars.
   - Update the backend code to read and use the new env vars.
3. **Refactor**:
   - Lint the full Go workspace: `cd go && make lint` from the workspace root.
   - Ensure no remaining hardcoded URLs in backend chimoney code (all should come from env vars).

## Acceptance Criteria

- [ ] `make all` in `local/` starts all services including `mockchimoney`
- [ ] Backend can connect to MockChimoney via `http://mockchimoney:8080`
- [ ] Traefik routes `mockchimoney.interledger.test` to the mock service
- [ ] Health check `/health` succeeds
- [ ] Full manual smoke test passes:
  - [ ] Create a new user account in Protea
  - [ ] Click "Personal Details" and complete the KYC widget
- [ ] All code passes `golangci-lint run ./..`
- [ ] Unit test coverage ≥ 75% (unit tests use memory-backed store); E2E tests use redis-backed store
- [ ] `make test` runs successfully (linting + unit tests + e2e tests)
  - [ ] Confirm KYC webhook is received and status updates
  - [ ] Click "Deposit" and complete a simulated Interac deposit via the pay page
  - [ ] Confirm deposit webhooks fire and balance increases
  - [ ] Click "Withdraw" and initiate an Interac withdrawal
  - [ ] Confirm withdrawal webhook fires and transaction completes
  - [ ] Temporal workflows completed successfully for each flow
- [ ] All previous feature scenarios still pass
- [ ] `cd go && make lint` passes with no errors
- [ ] No hardcoded URLs remain in backend chimoney code

## Manual Smoke Test Checklist

```
# Terminal 1: Start local environment
cd local
make all
# Wait ~30s for services to stabilize

# Terminal 2: Create user and run smoke test
cd local/scripts
./local-dev-tool rafiki --skip-ui --wait-for-ready 120

# Then in Protea browser:
# 1. Sign up new user
# 2. Complete KYC
# 3. Deposit $100 CAD via Interac
# 4. Withdraw $50 CAD via Interac
# 5. Check that all transactions appear in the UI and Temporal shows completed workflows
```
