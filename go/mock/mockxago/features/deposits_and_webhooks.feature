Feature: Xago Deposits and Webhook Notifications
  As a wallet backend
  I want to receive and process deposit notifications
  So that users' accounts are credited when they deposit funds

  Background:
    Given the Xago mock service is running
    And I have obtained a valid access token
    And I have created a sub-account for wallet "wallet_dep_test"
    And the wallet webhook URL is configured to "http://localhost:3000/xago/webhooks"

  Scenario: Simulate a test deposit
    When I simulate a test deposit with the following details:
      | accountId       | (the account id)    |
      | amount          | 5000.00             |
      | currencyCode    | ZAR                 |
      | depositReference | (the deposit ref)   |
    Then I receive a successful response with status code 200
    And the response includes a transaction ID
    And the response status is "pending"

  Scenario: Deposit webhook is sent after delay
    Given I have simulated a test deposit of 10000.00 ZAR
    When I wait 2 seconds for the webhook to be delivered
    Then the wallet receives a webhook with:
      | accountId       | (the account id)    |
      | amount          | 10000.00            |
      | currencyCode    | ZAR                 |
      | status          | completed           |
      | code            | 104                 |

  Scenario: Deposit webhook includes transaction ID
    Given I have simulated a test deposit
    When I wait for the webhook delivery
    Then the webhook includes:
      | transactionId   | (a valid UUID)      |
      | transactionReference | (the deposit reference) |

  Scenario: Webhook signature is valid
    Given I have simulated a test deposit
    When I wait for the webhook delivery
    Then the webhook includes valid headers:
      | x-gatehub-app-id | xago-mock           |
      | x-gatehub-timestamp | (current timestamp) |
      | x-gatehub-signature | (valid HMAC-SHA256) |

  Scenario: Deposit increases account balance
    Given the sub-account starts with zero ZAR balance
    When I simulate a test deposit of 5000.00 ZAR
    And I wait for the deposit to complete
    Then the sub-account ZAR balance is 5000.00

  @buggy
  Scenario: Multiple deposits accumulate correctly
    Given the sub-account starts with zero balance
    When I simulate a test deposit of 2000.00 ZAR
    And I wait for the deposit to complete
    And I simulate a test deposit of 3000.00 ZAR
    And I wait for the deposit to complete
    Then the total ZAR balance is 5000.00

  @buggy
  Scenario: Deposit for each currency is independent
    Given the sub-account has zero balance
    When I simulate a ZAR deposit of 1000.00
    And I wait for completion
    And I simulate a USD deposit of 500.00
    And I wait for completion
    Then the ZAR balance is 1000.00
    And the USD balance is 500.00
    And the balances are independent

  Scenario: List company deposits
    Given I have simulated 3 test deposits
    And all deposits have completed
    When I request to list company deposits with limit 10
    Then I receive a successful response with status code 200
    And the response includes 3 deposits
    And each deposit includes:
      | transactionId   | (valid UUID)        |
      | status          | completed           |
      | code            | 104                 |

  Scenario: List deposits with pagination
    Given I have simulated 15 test deposits
    When I request to list deposits with limit 10 and page 1
    Then the response includes 10 deposits
    When I request to list deposits with limit 10 and page 2
    Then the response includes 5 deposits

  Scenario: Deposit with correct reference routes to correct account
    Given I have created two sub-accounts with different deposit references
      | wallet_id | wallet_100  |
      | wallet_id | wallet_200  |
    When I simulate a ZAR deposit for wallet_100's deposit reference
    And I wait for completion
    Then wallet_100's balance is credited
    And wallet_200's balance remains unchanged

  Scenario: Reject deposit with invalid account ID
    When I attempt to simulate a deposit with invalid account ID "invalid_123":
      | amount          | 5000.00             |
      | currencyCode    | ZAR                 |
    Then I receive an error response with status code 400
    And the error message contains "account not found"

  Scenario: Reject deposit with zero amount
    When I attempt to simulate a deposit with amount 0.00
    Then I receive an error response with status code 400
    And the error message contains "amount must be greater than 0"

  Scenario: Reject deposit with negative amount
    When I attempt to simulate a deposit with amount -1000.00
    Then I receive an error response with status code 400
    And the error message contains "amount must be positive"

  Scenario: Deposit webhook payload matches expected format
    Given I have simulated a test deposit
    When I wait for the webhook delivery
    Then the webhook body includes all required fields:
      | accountId       | (valid UUID)        |
      | amount          | (the amount)        |
      | currencyCode    | (ZAR or USD)        |
      | transactionId   | (valid UUID)        |
      | status          | completed           |
      | code            | 104                 |
      | createdAt       | (ISO timestamp)     |
      | settledAt       | (ISO timestamp)     |

  Scenario: Deposit endpoint requires authentication
    Given I do not have a valid access token
    When I attempt to simulate a deposit without authentication
    Then I receive an error response with status code 401

  Scenario: List deposits requires authentication
    Given I do not have a valid access token
    When I attempt to list deposits without authentication
    Then I receive an error response with status code 401

  Scenario: Duplicate deposit with same reference is handled
    Given I have simulated a test deposit with specific deposit reference
    When I simulate another deposit with the same reference
    And I wait for both to complete
    Then both deposits are recorded
    And the account balance is credited twice
    And each deposit has a unique transaction ID

  Scenario: Deposits are tracked in company transactions
    Given I have simulated 5 test deposits
    When I wait for all deposits to complete
    And I request company deposits
    Then all 5 deposits appear in the list
    And each deposit shows as completed
