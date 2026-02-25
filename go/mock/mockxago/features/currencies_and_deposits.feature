Feature: Xago Currencies and Deposit Details
  As a wallet backend
  I want to retrieve available currencies and deposit details
  So that users can see where to send fiat deposits

  Background:
    Given the Xago mock service is running
    And I have obtained a valid access token

  Scenario: Retrieve list of available currencies
    When I request the list of available currencies
    Then I receive a successful response with status code 200
    And the response includes at least 2 currencies:
      | ZAR | South African Rand |
      | USD | US Dollar          |

  Scenario: Each currency includes bank account details
    When I request the list of available currencies
    Then I receive a successful response with status code 200
    And the ZAR currency includes:
      | currencyId    | ZAR             |
      | bankName      | FNB             |
      | accountNumber | (a valid number) |
      | branchCode    | 250145          |
      | swiftBIC      | FIRSZA22        |
    And the USD currency includes:
      | currencyId    | USD             |
      | bankName      | Citibank        |
      | accountNumber | (a valid number) |
      | branchCode    | 021             |
      | swiftBIC      | CITIUS33        |

  Scenario: Currency list is consistent across calls
    Given I have retrieved the currency list
    When I request the currency list again
    Then the response is identical to the previous response
    And all account numbers and bank codes remain the same

  Scenario: Get currency list without authentication
    Given I have not obtained an access token
    When I request the list of available currencies without authentication
    Then I receive a successful response with status code 200
    And the response includes available currencies

  Scenario: Sub-account creation includes matching bank details
    Given I have created a sub-account
    When I retrieve the created sub-account details
    Then the bankDepositDetails in the sub-account match the currencies endpoint
    And the ZAR bank details match exactly
    And the USD bank details match exactly

  Scenario: Deposit reference format is predictable
    Given I have created a sub-account for wallet with ID "wallet_test_123"
    When I retrieve the sub-account details
    Then the ZAR deposit reference contains "wallet_test_123"
    And the ZAR deposit reference contains "ZAR"
    And the USD deposit reference contains "wallet_test_123"
    And the USD deposit reference contains "USD"

  Scenario: Each sub-account gets unique deposit references
    Given I have created two sub-accounts
      | wallet_id | wallet_aaa |
      | wallet_id | wallet_bbb |
    When I retrieve both sub-account details
    Then the deposit references are different
    And deposit reference for wallet_aaa is unique
    And deposit reference for wallet_bbb is unique

  Scenario: Currency list returns nested banking providers (backend format)
    When I request the list of available currencies
    Then I receive a successful response with status code 200
    And the response includes currencies in nested format
    And the ZAR currency has nested banking providers with:
      | depositEnabled   | true            |
      | withdrawEnabled  | true            |
    And the first ZAR banking provider includes:
      | name              | FNB             |
      | depositAvailable  | true            |
    And the first ZAR provider deposit fields include:
      | bankName      | FNB             |
      | accountNumber | 62057334567     |
      | branchCode    | 250145          |
    And the USD currency has nested banking providers

  Scenario: Backend can parse nested currency response
    When I request the list of available currencies
    Then I receive a successful response with status code 200
    And the response structure matches backend expectations
    And each currency has required fields:
      | currencyCode     |
      | depositEnabled   |
      | bankingProviders |
    And each banking provider has required fields:
      | name             |
      | depositAvailable |
      | depositFields    |
