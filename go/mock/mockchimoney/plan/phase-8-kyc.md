# Phase 8 — KYC Widget

**Goal**: Browser-served KYC page with approve/decline, updating sub-account state and sending KYC webhooks.

**Definition of Done**: This phase is complete when `make test` passes successfully.

## Deliverables

- `internal/storage/interface.go` extended: `UpdateSubAccountKYCStatus`
- `internal/storage/memory.go` updated
- `internal/handler/kyc.go`:
  - `GET /verify/kyc/<externalID>?redirect=<url>` — serves `web/kyc.html`; 404 for unknown externalID; 400 if redirect missing
  - `POST /verify/kyc/<externalID>/approve` — sets KYC status to `completed`; redirects to `redirect` URL; enqueues `user.kyc.completed` webhook
  - `POST /verify/kyc/<externalID>/decline` — sets KYC status to `declined`; redirects with failure indicator; enqueues `user.kyc.declined` webhook
  - Returns 409 if KYC already completed
- `web/kyc.html` — minimal form with "Approve KYC" and "Decline KYC" buttons; hidden fields for `externalID` and `redirect`
- **Backend change**: Add `CHIMONEY_KYC_BASE_URL` env var support in `ops/ops.go` (see main plan.md section "Required Backend Changes")

## Feature File

`features/kyc.feature`

## Dependencies

- Phases 0–7 must be complete and green

## Test-Driven Development Notes

1. **Red**: Run KYC scenarios. Expect handler not found errors and missing HTML.
2. **Green**:
   - Implement handlers with basic KYC state management.
   - Serve `web/kyc.html` as embedded HTML.
   - Enqueue webhooks using the job queue from Phase 5.
3. **Refactor**:
   - Pay page (`web/pay.html`) and KYC page (`web/kyc.html`) both serve embedded HTML — extract a render helper.
   - Ensure webhook payloads match the expected structure in `features/kyc.feature`.
   - Run all previous feature scenarios.

## Acceptance Criteria

- [ ] KYC page is served for a valid sub-account ID
- [ ] KYC page returns 404 for unknown sub-account ID
- [ ] KYC page requires redirect query parameter
- [ ] Approving KYC redirects to the redirect URL
- [ ] Approving KYC updates sub-account status to `completed`
- [ ] `user.kyc.completed` webhook fires after approval
- [ ] Declining KYC redirects with a failure indicator
- [ ] Declining KYC updates sub-account status to `declined`
- [ ] `user.kyc.declined` webhook fires after rejection
- [ ] KYC can only be completed once (returns 409 on re-approval)
- [ ] Both KYC webhooks include valid svix signature headers
- [ ] All previous feature scenarios still pass
- [ ] All code passes `golangci-lint run ./...`
- [ ] Unit test coverage ≥ 60% (unit tests use memory-backed store); E2E tests use redis-backed store
- [ ] `make test` runs successfully (linting + unit tests + e2e tests)

## Backend Integration Notes

Coordinate with the backend team to add `CHIMONEY_KYC_BASE_URL` env var support:
- In `go/backend/providers/chimoney/ops/ops.go`, update `GetKYCWidget` to read the env var
- Default to `https://dash.chimoney.io` (prod) or `https://sandbox.chimoney.io` (non-prod)
- Local value should be `https://mockchimoney.interledger.test`
