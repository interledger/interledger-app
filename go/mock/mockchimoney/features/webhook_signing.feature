Feature: Webhook Signing and Delivery
  As the backend service
  I want outgoing webhooks to be signed with svix-style HMAC-SHA256
  So that they pass the backend's signature verification and cannot be replayed or forged

  # The backend verifies webhooks in ops/webhooks.go using the svix library.
  # Signature format: HMAC-SHA256 over "{svix-id}.{svix-timestamp}.{raw-body}"
  # using the base64-decoded suffix of CHIMONEY_WEBHOOK_SECRET (split on "_").
  # Example secret: "local_bG9jYWwtdGVzdC13ZWJob29rLXNlY3JldA=="
  #   → key = base64decode("bG9jYWwtdGVzdC13ZWJob29rLXNlY3JldA==")
  #         = "local-test-webhook-secret"

  Background:
    Given MockChimoney is running
    And I authenticate with a valid API key
    And a webhook receiver is listening
    And the configured webhook secret is "local_bG9jYWwtdGVzdC13ZWJob29rLXNlY3JldA=="

  # ── Svix header presence ──────────────────────────────────────────────────

  Scenario: Webhooks include the svix-id header
    Given I have triggered a deposit and waited for webhook delivery
    Then the received webhook includes the header "svix-id"
    And the "svix-id" value matches the pattern "msg_<uuid>"

  Scenario: Webhooks include the svix-timestamp header
    Given I have triggered a deposit and waited for webhook delivery
    Then the received webhook includes the header "svix-timestamp"
    And the "svix-timestamp" value is a Unix epoch integer

  Scenario: Webhooks include the svix-signature header
    Given I have triggered a deposit and waited for webhook delivery
    Then the received webhook includes the header "svix-signature"
    And the "svix-signature" value starts with "v1,"

  # ── Signature correctness ─────────────────────────────────────────────────

  Scenario: Webhook signature verifies correctly with the configured secret
    Given I have triggered a deposit and captured a webhook delivery
    When I verify the signature using the expected secret "local-test-webhook-secret"
    Then the signature is valid

  Scenario: Webhook signature fails verification with a different secret
    Given I have triggered a deposit and captured a webhook delivery
    When I verify the signature using the wrong secret "wrong-secret"
    Then the signature is invalid

  Scenario: Signature is computed over svix-id.svix-timestamp.body
    Given I have triggered a deposit and captured a webhook delivery
    When I manually compute HMAC-SHA256 over "{svix-id}.{svix-timestamp}.{raw-body}" with the configured key
    Then the result matches the base64 value in the "v1," prefix of "svix-signature"

  # ── Webhook payload structure ─────────────────────────────────────────────

  Scenario: Deposit webhook payload is a flat JSON object (no extra "data" wrapper)
    Given I have triggered a deposit and waited for webhook delivery
    Then the webhook body top-level fields include "eventType"
    And the webhook body top-level fields include "issueID"
    And the webhook body top-level fields include "status"
    And the webhook body does NOT contain a top-level "data" key wrapping the payload

  Scenario: Withdrawal webhook payload has the expected flat structure
    Given I have initiated a withdrawal and waited for webhook delivery
    Then the webhook body top-level fields include "eventType"
    And the webhook body top-level fields include "issueID"
    And the webhook body top-level fields include "meta"
    And the webhook body does NOT contain a top-level "data" key wrapping the payload

  Scenario: KYC webhook payload has the expected flat structure
    Given I have triggered KYC approval and waited for webhook delivery
    Then the webhook body top-level fields include "eventType"
    And the webhook body top-level fields include "userID"
    And the webhook body does NOT contain a top-level "data" key wrapping the payload

  # ── Secret format parsing ─────────────────────────────────────────────────

  Scenario: Secret format "prefix_base64payload" is parsed correctly
    Given the webhook secret is "myprefix_bXlzZWNyZXQ="
    When I trigger a deposit and capture a webhook
    Then the signature is valid when verified with key "mysecret"

  Scenario: Secret with multiple underscores in prefix is handled correctly
    Given the webhook secret is "multi_part_prefix_bXlzZWNyZXQ="
    When I trigger a deposit and capture a webhook
    Then the signature is valid when verified with key "mysecret"
