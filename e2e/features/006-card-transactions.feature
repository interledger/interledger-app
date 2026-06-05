@card-transactions
Feature: GateHub card transaction processing

  Background:
    Given a random test identifier is generated
    And the frontend is running at "https://interledger.test"
    And mockgatehub is running at "https://mockgatehub.interledger.test"
    And Rafiki assets are seeded
    And the card transaction test user is set up
    And the ledger balance for 'card-tx-user' should be total "200.00" EUR and available "200.00" EUR

  @card-transactions @withdrawal @purchase
  Scenario: Purchase transaction creates a pending withdrawal
    When I trigger a card purchase transaction for 'card-tx-user'
    Then a pending withdrawal transaction should exist in the database for 'card-tx-user' with note ""
    And the ledger balance for 'card-tx-user' should be total "200.00" EUR and available "198.11" EUR

  @card-transactions @withdrawal @atm
  Scenario: ATM withdrawal creates a pending transaction with ATM note
    When I trigger a card ATM withdrawal transaction for 'card-tx-user'
    Then a pending withdrawal transaction should exist in the database for 'card-tx-user' with note "ATM Withdrawal"
    And the ledger balance for 'card-tx-user' should be total "200.00" EUR and available "150.00" EUR

  @card-transactions @withdrawal @cash-advance
  Scenario: Cash advance creates a pending withdrawal with Cash Advance note
    When I trigger a card cash advance transaction for 'card-tx-user'
    Then a pending withdrawal transaction should exist in the database for 'card-tx-user' with note "Cash Advance"
    And the ledger balance for 'card-tx-user' should be total "200.00" EUR and available "180.00" EUR

  @card-transactions @withdrawal @preauth
  Scenario: Preauthorization creates a pending transaction with Preauthorization note
    When I trigger a card preauthorization transaction for 'card-tx-user'
    Then a pending withdrawal transaction should exist in the database for 'card-tx-user' with note "Preauthorization"
    And the ledger balance for 'card-tx-user' should be total "200.00" EUR and available "195.00" EUR

  @card-transactions @withdrawal @mastercard-send
  Scenario: Mastercard Send out creates a pending withdrawal with Mastercard Send note
    When I trigger a card transfer-from-account transaction for 'card-tx-user'
    Then a pending withdrawal transaction should exist in the database for 'card-tx-user' with note "Mastercard Send"
    And the ledger balance for 'card-tx-user' should be total "200.00" EUR and available "190.00" EUR

  @card-transactions @deposit @mastercard-send
  Scenario: Mastercard Send in creates a pending deposit with Mastercard Send note
    When I trigger a card transfer-to-account transaction for 'card-tx-user'
    Then a pending deposit transaction should exist in the database for 'card-tx-user' with note "Mastercard Send"
    And the ledger balance for 'card-tx-user' should be total "200.00" EUR and available "200.00" EUR

  @card-transactions @deposit @reversal
  Scenario: Purchase reversal creates a pending deposit with Refund note without touching the ledger
    When I trigger a card purchase transaction for 'card-tx-user'
    And I trigger a purchase reversal for the last card transaction of 'card-tx-user'
    Then a pending deposit transaction should exist in the database for 'card-tx-user' with note "Refund"
    And the first card transaction should be failed for 'card-tx-user'
    And the ledger balance for 'card-tx-user' should be total "200.00" EUR and available "198.11" EUR

  @card-transactions @withdrawal @preauth @incremental
  Scenario: Preauthorization incremental without reference creates a standalone pending withdrawal
    When I trigger a preauthorization incremental without reference for 'card-tx-user'
    Then a pending withdrawal transaction should exist in the database for 'card-tx-user' with note "Preauthorization"
    And the ledger balance for 'card-tx-user' should be total "200.00" EUR and available "195.00" EUR

  @card-transactions @withdrawal @preauth @incremental
  Scenario: Preauthorization incremental with reference rolls back the original and creates a new pending withdrawal
    When I trigger a card preauthorization transaction for 'card-tx-user'
    And I trigger a preauthorization incremental with reference to the last card transaction of 'card-tx-user'
    Then the first card transaction should be failed for 'card-tx-user'
    And a pending withdrawal transaction should exist in the database for 'card-tx-user' with note "Preauthorization"
    And the ledger balance for 'card-tx-user' should be total "200.00" EUR and available "195.00" EUR

  @card-transactions @informational @declined
  Scenario: Insufficient balance decline creates a failed transaction with Insufficient Balance note
    When I trigger a card insufficient-balance transaction for 'card-tx-user'
    Then a failed transaction should exist in the database for 'card-tx-user' with note "Insufficient Balance"
    And the ledger balance for 'card-tx-user' should be total "200.00" EUR and available "200.00" EUR

  @card-transactions @informational @declined
  Scenario: Wrong PIN creates a failed transaction with Invalid PIN note
    When I trigger a card wrong-pin transaction for 'card-tx-user'
    Then a failed transaction should exist in the database for 'card-tx-user' with note "Invalid PIN"
    And the ledger balance for 'card-tx-user' should be total "200.00" EUR and available "200.00" EUR

  @card-transactions @informational @declined
  Scenario: Invalid CVV creates a failed transaction with Invalid CVV note
    When I trigger a card invalid-cvv transaction for 'card-tx-user'
    Then a failed transaction should exist in the database for 'card-tx-user' with note "Invalid CVV"
    And the ledger balance for 'card-tx-user' should be total "200.00" EUR and available "200.00" EUR

  @card-transactions @informational @declined
  Scenario: Expired card creates a failed transaction with Card Expired note
    When I trigger a card expired transaction for 'card-tx-user'
    Then a failed transaction should exist in the database for 'card-tx-user' with note "Card Expired"
    And the ledger balance for 'card-tx-user' should be total "200.00" EUR and available "200.00" EUR

  @card-transactions @informational @declined
  Scenario: Frozen card creates a failed transaction with Frozen Card note
    When I trigger a frozen card transaction for 'card-tx-user'
    Then a failed transaction should exist in the database for 'card-tx-user' with note "Frozen Card"
    And the ledger balance for 'card-tx-user' should be total "200.00" EUR and available "200.00" EUR

  @card-transactions @informational
  Scenario: Card verification inquiry creates a completed transaction with Card Verification note
    When I trigger a card verification-inquiry transaction for 'card-tx-user'
    Then a completed transaction should exist in the database for 'card-tx-user' with note "Card Verification"
    And the ledger balance for 'card-tx-user' should be total "200.00" EUR and available "200.00" EUR

  @card-transactions @polling @clearing
  Scenario: Clearing poll finalizes a pending purchase transaction
    When I trigger a card purchase transaction for 'card-tx-user'
    And I set the card transaction status to "COMPLETED" for the last card transaction of 'card-tx-user'
    And I run the GateHub clearing card transactions poll workflow
    Then the last card transaction of 'card-tx-user' should be completed in the database
    And the ledger balance for 'card-tx-user' should be total "198.11" EUR and available "198.11" EUR

  @card-transactions @polling @realtime @reversal
  Scenario: Realtime poll finalizes a pending reversal deposit
    When I trigger a card purchase transaction for 'card-tx-user'
    And I trigger a purchase reversal for the last card transaction of 'card-tx-user'
    And I set the card transaction status to "COMPLETED" for the last card transaction of 'card-tx-user'
    And I run the GateHub realtime card transactions poll workflow
    Then the last card transaction of 'card-tx-user' should be completed in the database
    And the first card transaction should be failed for 'card-tx-user'
    And the ledger balance for 'card-tx-user' should be total "200.00" EUR and available "200.00" EUR

  @card-transactions @polling @realtime
  Scenario: Realtime poll finalizes a pending Mastercard Send transaction
    When I trigger a card transfer-from-account transaction for 'card-tx-user'
    And I set the card transaction status to "COMPLETED" for the last card transaction of 'card-tx-user'
    And I run the GateHub realtime card transactions poll workflow
    Then the last card transaction of 'card-tx-user' should be completed in the database
    And the ledger balance for 'card-tx-user' should be total "190.00" EUR and available "190.00" EUR

  @card-transactions @polling @clearing
  Scenario: Clearing poll rolls back a failed purchase transaction
    When I trigger a card purchase transaction for 'card-tx-user'
    And I set the card transaction status to "FAILED" for the last card transaction of 'card-tx-user'
    And I run the GateHub clearing card transactions poll workflow
    Then the last card transaction of 'card-tx-user' should be failed in the database
    And the ledger balance for 'card-tx-user' should be total "200.00" EUR and available "200.00" EUR

  @card-transactions @withdrawal @fx
  Scenario: Purchase with FX stores exchange rate data
    When I trigger a card purchase-with-fx transaction for 'card-tx-user'
    Then a pending withdrawal transaction should exist in the database for 'card-tx-user' with note ""
    And the last transaction for 'card-tx-user' should have exchange rate data
    And the ledger balance for 'card-tx-user' should be total "200.00" EUR and available "190.33" EUR

  @card-transactions @informational @fx
  Scenario: Declined transaction with FX stores exchange rate data
    When I trigger a card wrong-pin-with-fx transaction for 'card-tx-user'
    Then a failed transaction should exist in the database for 'card-tx-user' with note "Invalid PIN"
    And the last transaction for 'card-tx-user' should have exchange rate data
    And the ledger balance for 'card-tx-user' should be total "200.00" EUR and available "200.00" EUR
