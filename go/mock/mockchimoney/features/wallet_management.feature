Feature: Multicurrency Wallet Management
  As the backend service
  I want to create, retrieve, and transfer between chimoney sub-accounts
  So that each Interledger Wallet user has an associated chimoney wallet

  Background:
    Given MockChimoney is running
    And I authenticate with a valid API key

  # ── Create ───────────────────────────────────────────────────────────────

  Scenario: Create a wallet with only the required field
    When I POST /v0.2.4/multicurrency-wallets/create with body:
      | name | Alice |
    Then the response status is 201
    And the response JSON "status" is "success"
    And the response data contains an "id"
    And the response data "subAccount" is true

  Scenario: Create a wallet with all optional fields
    When I POST /v0.2.4/multicurrency-wallets/create with body:
      | name        | Bob Smith              |
      | email       | bob@example.com        |
      | firstName   | Bob                    |
      | lastName    | Smith                  |
      | phoneNumber | +15550001234           |
    Then the response status is 201
    And the response JSON "status" is "success"
    And the response data contains an "id"

  Scenario: Create a wallet without name is rejected
    When I POST /v0.2.4/multicurrency-wallets/create with body:
      | email | missing-name@example.com |
    Then the response status is 400
    And the response JSON "status" is "error"
    And the error message mentions "name"

  Scenario: Each wallet creation produces a unique ID
    When I create two wallets both named "Charlie"
    Then each wallet has a different "id" value

  Scenario: Newly created wallet has pending KYC verification status
    When I POST /v0.2.4/multicurrency-wallets/create with body:
      | name | David |
    Then the response data nested field "verification.status" is "pending"

  # ── Get ──────────────────────────────────────────────────────────────────

  Scenario: Retrieve a wallet by ID
    Given a wallet exists with name "Eve"
    When I GET /v0.2.4/multicurrency-wallets/get?id=<wallet ID>
    Then the response status is 200
    And the response JSON "status" is "success"
    And the response data "name" is "Eve"
    And the response data "subAccount" is true

  Scenario: Retrieve a non-existent wallet returns 404
    When I GET /v0.2.4/multicurrency-wallets/get?id=does-not-exist
    Then the response status is 404
    And the response JSON "status" is "error"

  Scenario: Retrieve a wallet without supplying id returns 400
    When I GET /v0.2.4/multicurrency-wallets/get without a query parameter
    Then the response status is 400
    And the response JSON "status" is "error"

  # ── Transfer ─────────────────────────────────────────────────────────────

  Scenario: Transfer between two existing sub-accounts
    Given a wallet exists for "Sender" with a known ID
    And a wallet exists for "Receiver" with a known ID
    When I POST /v0.2.4/multicurrency-wallets/transfer with body:
      | subAccount          | <sender ID>   |
      | receiver            | <receiver ID> |
      | amountToSend        | 50.00         |
      | originCurrency      | CAD           |
      | destinationCurrency | CAD           |
      | turnOffNotification | true          |
    Then the response status is 200
    And the response JSON "status" is "success"

  Scenario: Transfer requires amountToSend
    Given two wallets exist
    When I POST /v0.2.4/multicurrency-wallets/transfer without amountToSend
    Then the response status is 400
    And the error message mentions "amountToSend"

  Scenario: Transfer requires originCurrency
    Given two wallets exist
    When I POST /v0.2.4/multicurrency-wallets/transfer without originCurrency
    Then the response status is 400
    And the error message mentions "originCurrency"

  Scenario: Transfer requires destinationCurrency
    Given two wallets exist
    When I POST /v0.2.4/multicurrency-wallets/transfer without destinationCurrency
    Then the response status is 400
    And the error message mentions "destinationCurrency"

  Scenario: Transfer from non-existent sender returns 400
    Given a wallet exists for "Receiver" with a known ID
    When I POST /v0.2.4/multicurrency-wallets/transfer with:
      | subAccount          | non-existent-wallet-id |
      | receiver            | <receiver ID>          |
      | amountToSend        | 10.00                  |
      | originCurrency      | CAD                    |
      | destinationCurrency | CAD                    |
    Then the response status is 400
    And the response JSON "status" is "error"

  Scenario: The sendViaInterledger field is accepted and silently ignored
    Given two wallets exist
    When I POST /v0.2.4/multicurrency-wallets/transfer with sendViaInterledger true
    Then the response status is 200
