@stubbed
Feature: Xago Beneficiary Management
  As a wallet backend
  I want to manage bank account beneficiaries
  So that users can withdraw funds to their bank accounts

  Background:
    Given the Xago mock service is running
    And I have obtained a valid access token
    And I have created a sub-account for wallet "wallet_ben_test"

  Scenario: Add a new beneficiary successfully
    When I add a beneficiary with the following details:
      | name                 | My ABSA Account      |
      | scope                | external             |
      | currencyCode         | ZAR                  |
      | accountNumber        | 1234567890           |
      | branchCode           | 250155               |
      | bankName             | ABSA                 |
      | accountName          | John Doe             |
      | reference            | My ABSA Account      |
      | isOwn                | true                 |
    Then I receive a successful response with status code 200
    And the response includes a beneficiary with:
      | uuid        | (a valid UUID)        |
      | name        | My ABSA Account       |
      | currencyCode | ZAR                   |
      | status      | pending               |

  Scenario: Add beneficiary for USD account
    When I add a beneficiary with the following details:
      | name                 | My US Bank Account   |
      | scope                | external             |
      | currencyCode         | USD                  |
      | accountNumber        | 9876543210           |
      | branchCode           | 021                  |
      | bankName             | Citibank             |
      | accountName          | Jane Smith           |
      | reference            | My Citi Account      |
      | isOwn                | true                 |
    Then I receive a successful response with status code 200
    And the beneficiary is created with status "pending"

  Scenario: Beneficiary status transitions from pending to approved
    Given I have added a beneficiary with status "pending"
    When I wait 3 seconds for auto-approval
    Then I list the beneficiaries
    And the beneficiary status has transitioned to "approved"

  Scenario: Add beneficiary with minimal required fields
    When I add a beneficiary with only required fields:
      | name                 | Simple Account       |
      | currencyCode         | ZAR                  |
      | accountNumber        | 1111111111           |
      | bankName             | FNB                  |
      | accountName          | Test User            |
    Then I receive a successful response with status code 200
    And the beneficiary is created

  Scenario: Reject add beneficiary with missing name
    When I attempt to add a beneficiary without name:
      | currencyCode         | ZAR                  |
      | accountNumber        | 1234567890           |
      | branchCode           | 250155               |
      | bankName             | ABSA                 |
      | accountName          | John Doe             |
    Then I receive an error response with status code 400
    And the error message contains "name is required"

  Scenario: Reject add beneficiary with missing account number
    When I attempt to add a beneficiary without account number:
      | name                 | My ABSA Account      |
      | currencyCode         | ZAR                  |
      | branchCode           | 250155               |
      | bankName             | ABSA                 |
      | accountName          | John Doe             |
    Then I receive an error response with status code 400
    And the error message contains "accountNumber is required"

  Scenario: List beneficiaries for an account
    Given I have added 3 beneficiaries to the sub-account
    When I request to list beneficiaries with limit 10
    Then I receive a successful response with status code 200
    And the response includes 3 beneficiaries
    And each beneficiary includes required fields

  Scenario: List beneficiaries with pagination
    Given I have added 15 beneficiaries to the sub-account
    When I request to list beneficiaries with limit 10 and page 1
    Then I receive a successful response with status code 200
    And the response includes 10 beneficiaries
    And the pagination shows numberOfPages is 2
    When I request to list beneficiaries with limit 10 and page 2
    Then the response includes 5 beneficiaries

  Scenario: Each beneficiary gets a unique UUID
    Given I have added 2 beneficiaries
    When I list the beneficiaries
    Then the UUIDs are different
    And each UUID is unique

  Scenario: Beneficiaries are associated with correct sub-account
    Given I have created sub-accounts for two wallets
      | wallet_id | wallet_aaa |
      | wallet_id | wallet_bbb |
    And I have added beneficiaries to wallet_aaa
    When I list beneficiaries for wallet_aaa
    Then I get the correct beneficiaries
    And I do not get beneficiaries from wallet_bbb

  Scenario: Account beneficiary details are stored correctly
    When I add a beneficiary with specific details:
      | name                 | My Bank Account      |
      | currencyCode         | ZAR                  |
      | accountNumber        | 9876543210           |
      | branchCode           | 250145               |
      | bankName             | StandardBank         |
      | accountName          | John Doe             |
      | reference            | Reference_001        |
    And I list the beneficiaries
    Then the beneficiary details match exactly:
      | accountNumber        | 9876543210           |
      | branchCode           | 250145               |
      | bankName             | StandardBank         |
      | accountName          | John Doe             |
      | reference            | Reference_001        |

  Scenario: List beneficiaries without authentication
    Given I do not have a valid access token
    When I attempt to list beneficiaries without authentication
    Then I receive an error response with status code 401

  Scenario: Add beneficiary without authentication
    Given I do not have a valid access token
    When I attempt to add a beneficiary without authentication
    Then I receive an error response with status code 401
