# Phase 5 — Webhook Infrastructure and Deposit Webhooks

**Goal**: Svix-signed webhook delivery for deposits, backed by a simple in-process job queue.

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

- [ ] `charge.interac.completed` webhook fires after pay page completion
- [ ] `chimoney.redeem.completed` webhook fires after pay page completion
- [ ] Both webhooks are delivered in sequence
- [ ] Webhook issueID encodes the sub-account ID correctly
- [ ] Webhooks include valid `svix-id` header in format `msg_<uuid>`
- [ ] Webhooks include valid `svix-timestamp` header (Unix epoch)
- [ ] Webhooks include valid `svix-signature` header starting with `v1,`
- [ ] Webhook signature verifies correctly with the configured secret
- [ ] Webhook signature fails with wrong secret
- [ ] Webhook payload is a flat JSON object (no `data` wrapper)
- [ ] All previous feature scenarios still pass
- [ ] All code passes `golangci-lint run ./...`
- [ ] Unit test coverage ≥ 50% (unit tests use memory-backed store); E2E tests use redis-backed store
- [ ] `make test` runs successfully (linting + unit tests + e2e tests)

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
