Feature: Xago Sub-Account Management
  As a wallet backend
  I want to create and manage Xago sub-accounts
  So that users can have Xago accounts linked to their wallets

  Background:
    Given the Xago mock service is running
    And I have obtained a valid access token

  Scenario: Create a sub-account successfully
    When I create a sub-account with the following details:
      | firstName                 | John                                                    |
      | lastName                  | Doe                                                     |
      | email                     | john@example.com                                        |
      | mobileNumber              | +27123456789                                            |
      | identityType              | individual                                              |
      | idNumber                  | 9001011234567                                           |
      | physicalAddress           | 123 Main St, Cape Town, SA                              |
      | thirdPartyVerificationUrl | https://app.withpersona.com/dashboard/inquiries/inq_123 |
    Then I receive a successful response with status code 200
    And the sub-account is created with:
      | accountId               | (a valid UUID)                              |
      | depositAddress          | (a valid cryptocurrency address)            |
      | depositTag              | (a numeric value)                           |
    And the response includes bank deposit details for ZAR
    And the response includes bank deposit details for USD
    And the response includes beneficiaries with deposit references

  Scenario: Create sub-account with minimal required fields
    When I create a sub-account with only required fields:
      | firstName                 | Jane                                                    |
      | lastName                  | Smith                                                   |
      | email                     | jane@example.com                                        |
      | mobileNumber              | +27987654321                                            |
      | identityType              | individual                                              |
      | idNumber                  | 8512301234567                                           |
      | physicalAddress           | 456 Oak Ave, Johannesburg, SA                           |
      | thirdPartyVerificationUrl | https://app.withpersona.com/dashboard/inquiries/inq_456 |
    Then I receive a successful response with status code 200
    And a new sub-account is created
    And the sub-account has the provided email address

  Scenario: Reject sub-account creation with missing firstName
    When I attempt to create a sub-account without firstName:
      | lastName                  | Doe                                                     |
      | email                     | john@example.com                                        |
      | mobileNumber              | +27123456789                                            |
      | identityType              | individual                                              |
      | idNumber                  | 9001011234567                                           |
      | physicalAddress           | 123 Main St, Cape Town, SA                              |
      | thirdPartyVerificationUrl | https://app.withpersona.com/dashboard/inquiries/inq_123 |
    Then I receive an error response with status code 400
    And the error message contains "firstName is required"

  Scenario: Reject sub-account creation with missing lastName
    When I attempt to create a sub-account without lastName:
      | firstName                 | John                                                    |
      | email                     | john@example.com                                        |
      | mobileNumber              | +27123456789                                            |
      | identityType              | individual                                              |
      | idNumber                  | 9001011234567                                           |
      | physicalAddress           | 123 Main St, Cape Town, SA                              |
      | thirdPartyVerificationUrl | https://app.withpersona.com/dashboard/inquiries/inq_123 |
    Then I receive an error response with status code 400
    And the error message contains "lastName is required"

  Scenario: Reject sub-account creation with missing email
    When I attempt to create a sub-account without email:
      | firstName                 | John                                                    |
      | lastName                  | Doe                                                     |
      | mobileNumber              | +27123456789                                            |
      | identityType              | individual                                              |
      | idNumber                  | 9001011234567                                           |
      | physicalAddress           | 123 Main St, Cape Town, SA                              |
      | thirdPartyVerificationUrl | https://app.withpersona.com/dashboard/inquiries/inq_123 |
    Then I receive an error response with status code 400
    And the error message contains "email is required"

  Scenario: Sub-account includes deposit reference for routing
    When I create a sub-account with the following details:
      | firstName                 | Bob                                                     |
      | lastName                  | Johnson                                                 |
      | email                     | bob@example.com                                         |
      | mobileNumber              | +27111222333                                            |
      | identityType              | individual                                              |
      | idNumber                  | 8501121234567                                           |
      | physicalAddress           | 789 Pine Rd, Durban, SA                                 |
      | thirdPartyVerificationUrl | https://app.withpersona.com/dashboard/inquiries/inq_789 |
    Then I receive a successful response with status code 200
    And the beneficiaries in the response include:
      | beneficiaryType  | rollup                                    |
      | depositReference | (contains wallet ID and currency)         |
    And each currency has a unique deposit reference

  Scenario: Update sub-account with new verification URL
    Given I have created a sub-account for wallet "wallet_123"
    When I update the sub-account with new details:
      | thirdPartyVerificationUrl | https://app.withpersona.com/dashboard/inquiries/inq_999 |
      | idNumber                  | 9001011234567                                           |
      | physicalAddress           | 999 Updated St, Cape Town, SA                           |
    Then I receive a successful response with status code 200
    And the sub-account is updated with the new verification URL
    And the response contains updated status confirmation

  Scenario: Reject update with invalid accountId
    When I attempt to update a sub-account with invalid ID "invalid_id"
    Then I receive an error response with status code 400
    And the error message contains "invalid account ID format"

  Scenario: Sub-accounts are associated with wallet
    Given I have created two sub-accounts for different wallets
      | wallet_id  | wallet_abc |
      | wallet_id  | wallet_xyz |
    When I retrieve sub-account information for "wallet_abc"
    Then I get the correct sub-account associated with "wallet_abc"
    And I do not get sub-accounts from other wallets
