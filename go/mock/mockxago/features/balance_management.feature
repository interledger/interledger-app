Feature: Xago Balance Management
  As a wallet backend
  I want to manage currency balances for users
  So that I can track and display available funds

  Background:
    Given the Xago mock service is running
    And I have obtained a valid access token
    And I have created a sub-account for wallet "wallet_bal_test"

  Scenario: Get initial balance for new sub-account
    When I request the balance for the sub-account
    Then I receive a successful response with status code 200
    And the balance response includes:
      | accountId | (the created accountId) |
    And the balance includes ZAR currency with:
      | currencyCode | ZAR  |
      | available    | 0.00 |
      | reserved     | 0.00 |
      | total        | 0.00 |
    And the balance includes USD currency with:
      | currencyCode | USD  |
      | available    | 0.00 |
      | reserved     | 0.00 |
      | total        | 0.00 |

  Scenario: Balance includes both available and reserved amounts
    Given the sub-account has:
      | currency  | ZAR      |
      | available | 5000.00  |
      | reserved  | 500.00   |
    When I request the balance for the sub-account
    Then the balance response shows:
      | currencyCode | ZAR      |
      | available    | 5000.00  |
      | reserved     | 500.00   |
      | total        | 5500.00  |

  Scenario: Balance updates after deposit
    Given the sub-account starts with zero balance
    When a deposit of 10000.00 ZAR is received and processed
    Then I request the balance for the sub-account
    And the balance response shows:
      | currencyCode | ZAR       |
      | available    | 10000.00  |
      | total        | 10000.00  |

  Scenario: Balance updates after transfer
    Given the sub-account has a balance of:
      | currency | ZAR      |
      | amount   | 5000.00  |
    When a transfer of 2000.00 ZAR is initiated and completed
    Then I request the balance for the sub-account
    And the available ZAR balance is reduced to 3000.00
    And the total ZAR balance is reduced to 3000.00

  Scenario: Separate balances for different currencies
    Given the sub-account has:
      | currency | ZAR      |
      | amount   | 5000.00  |
    And the sub-account has:
      | currency | USD      |
      | amount   | 1000.00  |
    When I request the balance for the sub-account
    Then the ZAR balance is 5000.00
    And the USD balance is 1000.00
    And the balances are independent

  Scenario: Reject balance query with invalid account ID
    When I request the balance for an invalid account ID "invalid_123"
    Then I receive an error response with status code 400
    And the error message contains "invalid account ID"

  Scenario: Balance endpoint requires authentication
    Given I do not have a valid access token
    When I request the balance for a sub-account without authentication
    Then I receive an error response with status code 401

  Scenario: Multiple deposits accumulate correctly
    Given the sub-account starts with zero balance
    When a deposit of 1000.00 ZAR is received and processed
    And a deposit of 2000.00 ZAR is received and processed
    And a deposit of 1500.00 ZAR is received and processed
    Then I request the balance for the sub-account
    And the total ZAR balance is 4500.00

  Scenario: Balance accounts are linked to specific wallets
    Given I have created sub-accounts for two different wallets
      | wallet_id | wallet_100 |
      | wallet_id | wallet_200 |
    When I deposit 5000.00 ZAR to wallet_100
    Then the balance for wallet_100 is 5000.00
    And the balance for wallet_200 is 0.00
