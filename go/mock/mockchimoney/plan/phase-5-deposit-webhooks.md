# Phase 5 — Webhook Infrastructure and Deposit Webhooks

**Goal**: Svix-signed webhook delivery for deposits, backed by a simple in-process job queue.

**Status**: Completed (2026-03-20)

**Definition of Done**: This phase is complete when `make test` passes successfully.

## Deliverables

- `internal/webhook/sender.go` — `Send(ctx, webhookURL, secret, payload)`:
  - Parses `prefix_base64(key)` secret format (split on last `_`, base64-decode)
  - Generates `svix-id: msg_<uuid>`, `svix-timestamp: <unix-seconds>`
  - Signs `{svix-id}.{svix-timestamp}.{body}` with HMAC-SHA256
  - POSTs with all three headers
- `internal/jobs/queue.go` / `jobs/worker.go` — simple in-process channel-backed queue; each job is a delayed webhook send; configurable `WEBHOOK_MIN_DELAY_SEC`
- Wire job queue into `main.go`; enqueue two jobs on pay-page confirmation: `charge.interac.completed` then `chimoney.redeem.completed`

## Feature Files

- `features/deposits_and_webhooks.feature` (webhook scenarios)
- `features/webhook_signing.feature`

## Dependencies

- Phases 0–4 must be complete and green

## Test-Driven Development Notes

1. **Red**: Run feature scenarios for deposit webhooks. Expect no webhook delivery to the receiver.
2. **Green**:
   - Implement the job queue with a goroutine worker and channel.
   - Implement the webhook sender with proper svix signing.
   - Enqueuing jobs on payment confirmation should result in webhooks being sent.
3. **Refactor**:
   - Webhook sender must have unit tests that verify the signature exactly matches what the backend's svix `Verify()` would accept. Test this by computing the expected signature locally and comparing.
   - Extract secret parsing into its own function.
   - Ensure the job worker gracefully shuts down on context cancellation.
   - Run all previous feature scenarios.

## Acceptance Criteria

- [x] `charge.interac.completed` webhook fires after pay page completion
- [x] `chimoney.redeem.completed` webhook fires after pay page completion
- [x] Both webhooks are delivered in sequence
- [x] Webhook issueID encodes the sub-account ID correctly
- [x] Webhooks include valid `svix-id` header in format `msg_<uuid>`
- [x] Webhooks include valid `svix-timestamp` header (Unix epoch)
- [x] Webhooks include valid `svix-signature` header starting with `v1,`
- [x] Webhook signature verifies correctly with the configured secret
- [x] Webhook signature fails with wrong secret
- [x] Webhook payload is a flat JSON object (no `data` wrapper)
- [x] All previous feature scenarios still pass
- [x] All code passes `golangci-lint run ./...`
- [x] Unit test coverage ≥ 50% (unit tests use memory-backed store); E2E tests use redis-backed store
- [x] `make test` runs successfully (linting + unit tests + e2e tests)

## Implementation Notes

- Added svix webhook sender in `internal/webhook/sender.go` with `ParseSecret`, `ComputeSignature`, and `Send`.
- Added channel-backed queue and worker in `internal/jobs/job.go`, `internal/jobs/queue.go`, and `internal/jobs/worker.go`.
- Wired queue worker and sender in `cmd/mockchimoney/main.go`.
- Added deposit webhook enqueueing after pay confirmation in `internal/handler/payment.go`.
- Added webhook signature and payload tests in `internal/webhook/sender_test.go` and deposit webhook sequencing assertions in `internal/handler/payment_test.go`.

## Phase 6 Handoff Notes

- Reuse `generateIssueID` and webhook enqueue helpers from `internal/handler/common.go` and `internal/handler/webhook_enqueue.go` for payout webhook flow.

## Testing Notes

The webhook sender's signature function should be tested in isolation:
```go
// Example test
func TestSignatureGeneration(t *testing.T) {
    secret := "local_bG9jYWwtdGVzdC13ZWJob29rLXNlY3JldA=="
    // expectedKey = "local-test-webhook-secret" (base64-decoded)
    
    svixID := "msg_test123"
    timestamp := "1234567890"
    body := `{"eventType":"charge.interac.completed"}`
    
    sig := computeSignature(secret, svixID, timestamp, body)
    // Verify that sig can be validated by backend's svix.Verify()
}
```
