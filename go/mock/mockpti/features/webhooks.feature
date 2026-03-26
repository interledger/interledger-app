Feature: PTI webhook emissions to backend
  As backend PTI webhook consumer
  I want mockpti to emit webhook events
  So transaction and KYC callbacks are fully testable

  Scenario: Emit USER_ASSESSMENT webhook after accepted assessment
    Given webhook delivery is configured to backend "/webhooks/pti"
    And an existing PTI user assessment in state "ACCEPTED"
    When mockpti processes the assessment completion
    Then a webhook should be delivered with resource type "USER_ASSESSMENT"
    And the webhook payload should include user id and request id

  Scenario: Emit TRANSACTION_STATUS webhook for settled deposit
    Given webhook delivery is configured to backend "/webhooks/pti"
    And an existing PTI deposit transaction in state "PENDING"
    When mockpti transitions the transaction to "SETTLED"
    Then a webhook should be delivered with resource type "TRANSACTION_STATUS"
    And the webhook payload should include transaction type "DEPOSIT"
    And the webhook payload should include status "SETTLED"

  Scenario: Emit TRANSACTION_STATUS webhook for failed withdrawal
    Given webhook delivery is configured to backend "/webhooks/pti"
    And an existing PTI withdrawal transaction in state "PENDING"
    When mockpti transitions the transaction to "REFUSED"
    Then a webhook should be delivered with resource type "TRANSACTION_STATUS"
    And the webhook payload should include transaction type "WITHDRAWAL"
    And the webhook payload should include status "REFUSED"

  Scenario: Emit encrypted and signed USER_ASSESSMENT webhook
    Given webhook delivery is configured to backend "/webhooks/pti"
    And an existing PTI user assessment in state "ACCEPTED"
    When mockpti processes the assessment completion
    Then a webhook should be delivered with resource type "USER_ASSESSMENT"
    And the webhook payload should be signed and encrypted
    And the webhook payload should include user id and request id
