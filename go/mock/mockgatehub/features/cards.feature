@cards
Feature: Card management and lifecycle
  As a wallet integrator using cards
  I want to manage card customers and their cards
  So that I can provide card functionality to users

  Background:
    Given a clean MockGatehub instance
    And a managed user with KYC state "accepted"
    And HMAC headers using app id "local-test-app-id" and secret "local-test-app-secret"

  Scenario: Create a managed customer with EUR account and initial card
    When I POST /cards/v1/customers/managed with managed user UUID header, nameOnCard "John Doe", account productCode "PWSR_DEBP_2404", currency "EUR", and card productCode "PWSR_DEBP_2404"
    Then the response status is 201
    And the response contains a customer with id, sourceId, type "Citizen", and kycStatus "accepted"
    And the customer has one account with currency "EUR", status "ACTIVE", and type "DEBIT"
    And the account has one card with status "Active", nameOnCard "John Doe", and expiryDate in the future

  Scenario: List cards for a customer
    Given a managed customer with a card exists
    When I GET /cards/v1/cards/{customerId}?pageSize=100 with managed user UUID header
    Then the response status is 200
    And the response has a data array with at least one card
    And each card has id, accountId, customerId, nameOnCard, maskedPan, status, expiryDate, and productCode
    And the response includes pagination with pageNumber, pageSize, and totalPages

  Scenario: Get card details
    Given a managed customer with a card exists
    When I GET /cards/v1/cards/{cardId}/card with managed user UUID header
    Then the response status is 200
    And the response contains the card object with all required fields
    And the card status is "Active"

  Scenario: Lock card temporarily
    Given a managed customer with a card exists
    When I PUT /cards/v1/cards/{cardId}/lock?reasonCode=ClientRequestedLock with managed user UUID header and note "Testing lock"
    Then the response status is 200
    And the card status is changed to "TemporaryBlocked"

  Scenario: Unlock card
    Given a managed customer with a locked card exists
    When I PUT /cards/v1/cards/{cardId}/unlock with managed user UUID header and note "Testing unlock"
    Then the response status is 200
    And the card status is changed back to "Active"

  Scenario: Close/delete card
    Given a managed customer with a card exists
    When I DELETE /cards/v1/cards/{cardId}/card?reasonCode=UserRequest with managed user UUID header
    Then the response status is 200
    And the card no longer appears in list cards query (filtered out)

  Scenario: Get card token for secure data access
    Given a managed customer with a card exists
    When I POST /cards/v1/token/card-data with cardId and managed user UUID header
    Then the response status is 200
    And the response contains a token starting with "mock-card-data-"
    And the response contains a links array with at least one entry

  Scenario: Get card limits
    Given a managed customer with a card exists
    When I GET /cards/v1/cards/{cardId}/limits with managed user UUID header
    Then the response status is 200
    And the response is an array of card limit objects
    And each limit has type, limit amount, currency "EUR", and isDisabled flag
    And the limits include types: dailyOverall, perTransaction, monthlyOverall, dailyAtm, dailyEcomm

  Scenario: Set card limits
    Given a managed customer with a card exists
    When I PUT /cards/v1/cards/{cardId}/limits with managed user UUID header and updated dailyOverall limit 2000
    Then the response status is 200
    And the response returns the updated limits array
    And the dailyOverall limit is changed to 2000

  Scenario: Create card transaction
    Given a managed customer with a card exists
    When I POST /cards/v1/transactions with cardId, amount "25.00", currency "EUR", and managed user UUID header
    Then the response status is 201
    And the response contains a card transaction with transactionId and ghResponseCode "00"

  Scenario: Get card transaction details
    Given a managed customer with a card and transaction exists
    When I GET /cards/v1/transactions/{transactionId} with managed user UUID header
    Then the response status is 200
    And the response contains the card transaction with transactionId and transactionAmount

  Scenario: Get pending 3DS confirmations
    Given a managed customer with a pending 3DS challenge exists
    When I GET /cards/v1/transaction/pending-confirmations with managed user UUID header
    Then the response status is 200
    And the response contains pending 3DS confirmations
    And each confirmation includes transactionId, merchantName, purchaseAmount, purchaseCurrency, and timeout

  Scenario: Confirm 3DS payment (approve)
    Given a managed customer with a pending 3DS challenge exists
    When I POST /cards/v1/transaction/{transactionId} with managed user UUID header, confirmed true, authMethod "pin"
    Then the response status is 200
    And the response indicates success with status "approved"

  Scenario: Decline 3DS payment
    Given a managed customer with a pending 3DS challenge exists
    When I POST /cards/v1/transaction/{transactionId} with managed user UUID header, confirmed false, authMethod "pin"
    Then the response status is 200
    And the response indicates declined with status "declined"

  Scenario: Get card application products
    When I GET /cards/v1/card-applications/test-app-id/card-products with managed user UUID header
    Then the response status is 200
    And the response contains a data array with card product objects

  Scenario: Order plastic card for a card
    Given a managed customer with a card exists
    When I POST /cards/v1/cards/{cardId}/plastic with managed user UUID header
    Then the response status is 201
    And the response contains orderId, cardId, status "PENDING", and type "PLASTIC"
