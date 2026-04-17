Feature: PTI transactions for deposit, withdrawal, and transfer
  As backend payment workflows
  I want PTI transaction endpoints with realistic states
  So onboarding, offboarding, and P2P flows are testable

  Scenario: Create deposit transaction
    Given an existing PTI user with a USD wallet and bank account
    And valid PTI headers are present
    When I POST "/transactions/deposits" with a valid deposit payload
    Then the response status should be 200
    And the response should include a transaction request id

  Scenario: Create withdrawal transaction
    Given an existing PTI user with a USD wallet and bank account
    And valid PTI headers are present
    When I POST "/transactions/withdrawals" with a valid withdrawal payload
    Then the response status should be 200
    And the response should include a transaction request id

  Scenario: Create transfer transaction
    Given two PTI users each with a USD wallet
    And valid PTI headers are present
    When I POST "/transactions/transfers" with a valid transfer payload
    Then the response status should be 200
    And the response should include a transaction request id

  Scenario: Get transaction status
    Given an existing PTI transaction request id
    When I GET "/transactions/{requestId}"
    Then the response status should be 200
    And the response should include "status"

  Scenario: Update transaction status with feedback
    Given an existing PTI transaction request id
    And valid PTI headers are present
    When I POST "/transactions/{requestId}/updates" with feedback payload
    Then the response status should be 200
    And the response should include an update id

  Scenario: Wallet balance goes negative after settled deposit is returned following withdrawal
    Given webhook delivery is configured to backend "/webhooks/pti"
    And an existing PTI user with a USD wallet and bank account
    When mockpti settles a deposit
    And mockpti settles a withdrawal
    And mockpti returns the deposit
    Then a webhook should be delivered with resource type "TRANSACTION_STATUS"
    And the webhook payload should include transaction type "DEPOSIT"
    And the webhook payload should include status "RETURNED"
    And the wallet balance should be negative

  Scenario: Wallet balance is restored to deposit amount after a returned withdrawal
    Given webhook delivery is configured to backend "/webhooks/pti"
    And an existing PTI user with a USD wallet and bank account
    When mockpti settles a deposit
    And mockpti settles a withdrawal
    And mockpti returns the withdrawal
    Then a webhook should be delivered with resource type "TRANSACTION_STATUS"
    And the webhook payload should include transaction type "WITHDRAWAL"
    And the webhook payload should include status "RETURNED"
    And the wallet balance should equal the deposited amount
