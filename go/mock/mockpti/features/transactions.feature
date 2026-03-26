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
