Feature: API Key Authentication
  As the backend service
  I want MockChimoney to validate X-API-KEY on protected endpoints
  So that only authorised callers can create wallets or initiate payments

  Background:
    Given MockChimoney is running with authentication enforced
    And the configured API key is "local-test-api-key"

  Scenario: Valid API key is accepted
    When I POST /v0.2.4/multicurrency-wallets/create with header "X-API-KEY: local-test-api-key" and body:
      | name | Test Wallet |
    Then the response status is 201
    And the response JSON "status" is "success"

  Scenario: Missing X-API-KEY header is rejected
    When I POST /v0.2.4/multicurrency-wallets/create without an X-API-KEY header and body:
      | name | Test Wallet |
    Then the response status is 401
    And the response JSON "status" is "error"

  Scenario: Wrong API key is rejected
    When I POST /v0.2.4/multicurrency-wallets/create with header "X-API-KEY: wrong-key" and body:
      | name | Test Wallet |
    Then the response status is 401
    And the response JSON "status" is "error"

  Scenario: API key is also required on payment endpoints
    When I POST /v0.2.4/payment/initiate without an X-API-KEY header and body:
      | amount     | 100.00            |
      | currency   | CAD               |
      | subAccount | some-sub-account  |
      | payerEmail | payer@example.com |
    Then the response status is 401

  Scenario: API key is also required on payout endpoints
    When I POST /v0.2.4/payouts/interac without an X-API-KEY header and body:
      | debitCurrency | CAD |
    Then the response status is 401

  Scenario: Authentication can be disabled for development convenience
    Given MockChimoney is running with authentication disabled
    When I POST /v0.2.4/multicurrency-wallets/create without an X-API-KEY header and body:
      | name | Test Wallet |
    Then the response status is 201

  Scenario: Health check does not require authentication even when enforced
    When I GET /health without an X-API-KEY header
    Then the response status is 200
