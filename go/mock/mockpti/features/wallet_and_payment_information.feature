Feature: PTI wallets and payment information
  As backend PTI integration
  I want to create wallets and bank payment information
  So USD balance and bank-account linking flows work in tests

  Scenario: Create a PTI wallet for a user
    Given an existing PTI user in mockpti
    And valid PTI headers are present
    When I POST "/users/{userId}/wallets" with USD wallet payload
    Then the response status should be 200
    And the response should include a wallet id

  Scenario: List and fetch PTI wallets
    Given an existing PTI user with at least one wallet in mockpti
    When I GET "/users/{userId}/wallets"
    Then the response status should be 200
    And the response should include at least one wallet
    When I GET "/users/{userId}/wallets/{walletId}"
    Then the response status should be 200
    And the response should include "walletId" equal to "{walletId}"

  Scenario: Create and read bank account payment information
    Given an existing PTI user in mockpti
    And valid PTI headers are present
    When I POST "/users/{userId}/payment-information" with bank account payload
    Then the response status should be 200
    And the response should include a payment information id
    When I GET "/users/{userId}/payment-information/{paymentInformationId}"
    Then the response status should be 200
    And the response should include "type" equal to "BANK_ACCOUNT"
