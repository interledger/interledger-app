Feature: Xago Integration Workflows
  As a wallet backend
  I want to perform complete user workflows
  So that users can deposit, transfer, and withdraw funds

  Background:
    Given the Xago mock service is running
    And the wallet backend is configured to use Xago
    And the webhook receiver is listening

  Scenario: Complete onboarding workflow for ZA user
    Given a user completes KYC for a South African wallet
    When the KYC status is marked as approved
    Then the wallet backend automatically:
      | step 1 | Calls Xago to create a sub-account |
      | step 2 | Creates a ZAR balance account       |
      | step 3 | Links Persona verification URL      |
    And the user can see deposit details in the wallet UI

  Scenario: User receives a deposit and transfers funds
    Given a user has completed onboarding
    And the user has access to their deposit details
    When the user receives a deposit of 10000.00 ZAR
      | method | Bank transfer to provided account |
    And I simulate the deposit with test deposit endpoint
    Then the wallet balance increases by 10000.00 ZAR
    When the user adds a beneficiary bank account
    And the beneficiary is auto-approved
    And the user initiates a transfer of 2000.00 ZAR
    Then the balance is reduced by 2000.00
    And the remaining balance is 8000.00

  Scenario: User manages multiple beneficiaries
    Given a user has an active ZAR balance account
    When the user adds beneficiary #1
      | name | My Home Bank Account |
      | bank | FNB                  |
    And the user adds beneficiary #2
      | name | My Work Bank Account |
      | bank | ABSA                 |
    Then both beneficiaries are listed
    And both beneficiaries are auto-approved
    When the user transfers to beneficiary #1
    And the user transfers to beneficiary #2
    Then both transfers complete successfully
    And the balance is debited twice

  Scenario: User manages both ZAR and USD accounts
    Given a user has completed ZAR onboarding
    And the user has a ZAR balance account
    When the user manually creates a USD balance account
    Then the user has both ZAR and USD balance accounts
    When the user receives a ZAR deposit of 5000.00
    And the user receives a USD deposit of 1000.00
    And the user transfers 1000.00 ZAR to a beneficiary
    And the user transfers 500.00 USD to a beneficiary
    Then ZAR balance is 4000.00
    And USD balance is 500.00
    And the balances are independent

  Scenario: Token refresh during long operation
    Given a user is in the middle of a wallet session
    And their access token is about to expire
    When they request a new token
    Then they receive a new valid token
    When they use the new token for subsequent operations
    Then all operations succeed
    And their session continues uninterrupted

  Scenario: Error handling in deposit flow
    Given a user attempts to receive a deposit
    When the webhook delivery fails the first time
    Then the mock retries the webhook delivery
    And the deposit is eventually credited
    And the account balance reflects the final amount

  Scenario: Concurrent transfers don't cause double-spending
    Given a user has a balance of 5000.00 ZAR
    And the user has two beneficiaries
    When the user initiates two concurrent transfers
      | transfer 1 | 3000.00 ZAR to beneficiary 1 |
      | transfer 2 | 3000.00 ZAR to beneficiary 2 |
    Then one transfer succeeds
    And one transfer fails with insufficient balance error
    And the balance is only reduced once
    And the remaining balance is 2000.00

  Scenario: Deposit reference enables correct routing
    Given two users have completed onboarding
      | user 1 | wallet_aaa |
      | user 2 | wallet_bbb |
    And each has a unique deposit reference
    When funds are deposited to the shared Xago account using user 1's reference
    Then user 1's balance increases
    And user 2's balance does not change
    When funds are deposited using user 2's reference
    Then user 2's balance increases
    And user 1's balance only includes the first deposit

  Scenario: Wallet state consistency after failed operation
    Given a user has a balance of 10000.00 ZAR
    When the user attempts an operation that fails
      | operation | Transfer with invalid beneficiary |
    Then the balance remains 10000.00
    And no transaction is created
    And the account state is unchanged

  Scenario: Complete audit trail of transactions
    Given a user completes the following operations:
      | step 1 | Receive 10000.00 ZAR deposit      |
      | step 2 | Receive 2000.00 ZAR deposit       |
      | step 3 | Transfer 5000.00 ZAR              |
      | step 4 | Transfer 1000.00 ZAR              |
    When I request the transaction history
    Then all 4 transactions are recorded
    And each transaction includes complete details
    And the transactions are in chronological order
    And the final balance calculation is correct

  Scenario: Sub-account creation includes all necessary details
    Given the wallet backend initiates sub-account creation
    When the sub-account is created with user KYC data
    Then the response includes:
      | bank details for ZAR |
      | bank details for USD |
      | deposit references   |
      | deposit addresses    |
    And the user can immediately see all required information
    And the user can start receiving deposits

  Scenario: Beneficiary auto-approval flow
    Given a user has an active sub-account
    When the user adds a new beneficiary
    Then the beneficiary status is initially "pending"
    When I wait 3 seconds for auto-approval
    And I list the beneficiaries again
    Then the beneficiary status is "approved"
    And the user can immediately transfer to this beneficiary
