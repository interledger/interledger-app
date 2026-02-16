@stubbed
Feature: Xago Transactions and Transfers
  As a wallet backend
  I want to create and track transactions
  So that users can transfer funds to beneficiaries

  Background:
    Given the Xago mock service is running
    And I have obtained a valid access token
    And I have created a sub-account for wallet "wallet_txn_test"
    And the sub-account has a ZAR balance of 10000.00
    And I have added a ZAR beneficiary

  Scenario: Create a transfer successfully
    When I create a transfer with the following details:
      | amount          | 1000.50              |
      | currencyCode    | ZAR                  |
      | beneficiaryId   | (the beneficiary id) |
      | reference       | Invoice #123         |
      | idempotencyKey  | (a unique key)       |
    Then I receive a successful response with status code 200
    And the response contains a transaction ID
    And the transaction ID is a valid UUID

  Scenario: Transfer deducts from account balance
    Given the sub-account has a balance of 5000.00 ZAR
    When I create a transfer of 1000.00 ZAR
    Then the transfer is successful
    And the new balance for ZAR is 4000.00

  Scenario: Transfer with insufficient balance fails
    Given the sub-account has a balance of 500.00 ZAR
    When I attempt to create a transfer of 1000.00 ZAR
    Then I receive an error response with status code 400
    And the error message contains "insufficient balance"
    And the balance remains 500.00

  Scenario: Transfer auto-completes after delay
    When I create a transfer of 500.00 ZAR
    And I wait 3 seconds for processing
    Then I retrieve the transaction
    And the transaction status is "completed"

  Scenario: Idempotent transfer requests
    Given I create a transfer with idempotency key "key_001":
      | amount          | 1000.00              |
      | currencyCode    | ZAR                  |
      | beneficiaryId   | (beneficiary id)     |
    When I create another transfer with the same idempotency key "key_001"
    Then I receive the same transaction ID
    And only one transfer is deducted from the balance
    And the balance is deducted once, not twice

  Scenario: Reject transfer without beneficiary ID
    When I attempt to create a transfer without beneficiary ID:
      | amount          | 1000.00              |
      | currencyCode    | ZAR                  |
      | reference       | Invoice #123         |
    Then I receive an error response with status code 400
    And the error message contains "beneficiaryId is required"

  Scenario: Reject transfer with invalid beneficiary ID
    When I attempt to create a transfer with invalid beneficiary ID "invalid_ben_123":
      | amount          | 1000.00              |
      | currencyCode    | ZAR                  |
      | reference       | Invoice #123         |
    Then I receive an error response with status code 400
    And the error message contains "beneficiary not found"

  Scenario: Reject transfer with zero amount
    When I attempt to create a transfer with amount 0.00
    Then I receive an error response with status code 400
    And the error message contains "amount must be greater than 0"

  Scenario: Reject transfer with negative amount
    When I attempt to create a transfer with amount -500.00
    Then I receive an error response with status code 400
    And the error message contains "amount must be positive"

  Scenario: List transactions for an account
    Given I have created 3 transfers
    When I request to list transactions with limit 10
    Then I receive a successful response with status code 200
    And the response includes 3 transactions
    And each transaction includes required fields

  Scenario: List transactions with pagination
    Given I have created 15 transfers
    When I request to list transactions with limit 10 and page 1
    Then the response includes 10 transactions
    When I request to list transactions with limit 10 and page 2
    Then the response includes 5 transactions

  Scenario: Retrieve details for a single transaction
    Given I have created a transfer with transaction ID "txn_123"
    When I retrieve transaction details for "txn_123"
    Then I receive a successful response with status code 200
    And the response includes the transfer details:
      | transactionId   | txn_123              |
      | status          | completed            |
      | amount          | (the transferred amount) |
      | currencyCode    | ZAR                  |

  Scenario: Transaction includes created and settled timestamps
    When I create a transfer of 1000.00 ZAR
    And I wait for the transfer to complete
    Then I retrieve the transaction
    And the transaction includes:
      | createdAt   | (a valid timestamp)  |
      | settledAt   | (a valid timestamp)  |
    And the settledAt timestamp is after createdAt

  Scenario: Transfers are associated with correct beneficiary
    Given I have created 2 beneficiaries
    And I have created transfers to both beneficiaries
    When I list transactions
    Then each transaction references the correct beneficiary ID
    And transfers are not mixed between beneficiaries

  Scenario: Transaction requires authentication
    Given I do not have a valid access token
    When I attempt to create a transfer without authentication
    Then I receive an error response with status code 401

  Scenario: List transactions requires authentication
    Given I do not have a valid access token
    When I attempt to list transactions without authentication
    Then I receive an error response with status code 401

  Scenario: Multiple transfers from same account
    Given the sub-account has a balance of 10000.00 ZAR
    When I create a transfer of 2000.00 ZAR
    And I create a transfer of 3000.00 ZAR
    And I create a transfer of 1000.00 ZAR
    Then the total amount transferred is 6000.00
    And the remaining balance is 4000.00
    And I can retrieve all 3 transactions

  Scenario: Transfer to different currencies
    Given the sub-account has:
      | currency | ZAR  |
      | amount   | 5000.00 |
    And the sub-account has:
      | currency | USD  |
      | amount   | 1000.00 |
    And I have added both ZAR and USD beneficiaries
    When I create a ZAR transfer of 1000.00
    And I create a USD transfer of 200.00
    Then the ZAR balance is reduced to 4000.00
    And the USD balance is reduced to 800.00
