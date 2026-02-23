Feature: Xago Integration Workflows
  As a wallet backend
  I want to perform complete user workflows
  So that users can deposit, transfer, and withdraw funds

  Background:
    Given the Xago mock service is running
    And the wallet backend is configured to use Xago
    And I have obtained a valid access token

  Scenario: Complete onboarding via KYC iframe submission
    Given I create a sub-account with the following details:
      | walletId  | wallet_kyc_001    |
      | firstName | Sarah             |
      | lastName  | Smith             |
      | email     | sarah@example.com |
    When I submit KYC for wallet "wallet_kyc_001" with name "Sarah" "Smith"
    Then the KYC submission is accepted
    And the sub-account for "wallet_kyc_001" is retrievable

  Scenario: Deposit and transfer with correct balance updates
    Given I create a sub-account with the following details:
      | walletId  | wallet_flow_001  |
      | firstName | John             |
      | lastName  | Doe              |
      | email     | john@example.com |
    And the sub-account starts with zero balance
    When a deposit of 10000.00 ZAR is received and processed
    And I request the balance for the sub-account
    Then the ZAR balance is 10000.00
    When a transfer of 2000.00 ZAR is initiated and completed
    And I request the balance for the sub-account
    Then the ZAR balance is 8000.00

  Scenario: Multi-currency balances are tracked independently
    Given I create a sub-account with the following details:
      | walletId  | wallet_multi_001  |
      | firstName | Alice             |
      | lastName  | Jones             |
      | email     | alice@example.com |
    And the sub-account starts with zero balance
    When a deposit of 5000.00 ZAR is received and processed
    And a deposit of 1000.00 USD is received and processed
    And I request the balance for the sub-account
    Then the ZAR balance is 5000.00
    And the USD balance is 1000.00
    And the balances are independent

  Scenario: Deposit reference routing isolates funds per wallet
    Given I have created a sub-account for wallet "wallet_aaa"
    And I have created a sub-account for wallet "wallet_bbb"
    When I deposit 5000.00 ZAR to wallet_aaa
    And I deposit 2000.00 ZAR to wallet_bbb
    Then the balance for wallet_aaa is 5000.00
    And the balance for wallet_bbb is 2000.00

  Scenario: Access token can be refreshed and reused
    Given I have obtained an access token that is about to expire
    When I request a new login token with valid credentials
    Then the response contains a valid access token
    And the new token is different from the expired token
    And the new token is valid

  Scenario: Transaction history records created transactions
    Given I create a sub-account with the following details:
      | walletId  | wallet_audit_001 |
      | firstName | Bob              |
      | lastName  | Brown            |
      | email     | bob@example.com  |
    When I create the following test transactions:
      | 10000.00 | ZAR |
      |  2000.00 | ZAR |
      |  5000.00 | ZAR |
      |  1000.00 | USD |
    Then the transaction history contains at least 4 records

  Scenario: Sub-account creation includes all deposit details
    Given I create a sub-account with the following details:
      | walletId  | wallet_details_001 |
      | firstName | Eve                |
      | lastName  | Black              |
      | email     | eve@example.com    |
    Then the sub-account is created with:
      | accountId      | (auto-generated) |
      | depositAddress | (auto-generated) |
      | depositTag     | (auto-generated) |
    And the response includes bank deposit details for ZAR
    And the response includes bank deposit details for USD
    And the response includes beneficiaries with deposit references
