Feature: Transaction handling
  As an application consuming MockGatehub
  I want deposits and hosted transfers to complete predictably
  So that payment and payout flows do not stall

  Background:
    Given a managed user with at least one wallet address
    And authenticated requests signed with HMAC headers

  Scenario: Complete a dynamic deposit via iframe token
    Given an iframe token obtained with scope ["deposit"] and mapped to the managed user
    When I POST /transaction/complete?paymentType=deposit&bearer={iframeToken} with amount "75.50" and currency "EUR" and Authorization header "Bearer {iframeToken}"
    Then the response status is 200 with status "success"
    And the user balance for EUR increases by 75.50

  Scenario: Hosted transfer returns completed transaction and webhook
    When I POST /core/v1/transactions with user_id, amount 150.00, currency "USD", type 2, and deposit_type "hosted"
    Then the response includes an id or uuid
    And the amount echoes 150.00
    And the status is completed (integer 1 or string "completed")
    And deposit_type is "hosted"
    And a core.deposit.completed webhook is emitted for the user

  Scenario: External deposit transaction formatting
    When I POST /core/v1/transactions with type 1, deposit_type "external", receiving_address set to a wallet, amount 123.45, currency "EUR", and a valid vault_uuid
    Then the response status is 201
    And fields amount, total_amount, and fee are string formatted with two decimals
    And status is integer 1
    And the transaction can be retrieved via GET /core/v1/transactions/{id} with the same format

  Scenario: Hosted transfer with sending_address debits the sender
    Given the user has a funded EUR wallet with balance 500.00
    When I POST /core/v1/transactions with sending_address as the user wallet, receiving_address as "settlement", amount 200.00, currency "EUR", type 2, deposit_type "hosted"
    Then the response status is 201
    And the user balance for EUR is 300.00

  Scenario: Hosted transfer with receiving_address credits the receiver
    Given the user has a funded EUR wallet with balance 500.00
    When I POST /core/v1/transactions with receiving_address as the user wallet, sending_address as "settlement", amount 200.00, currency "EUR", type 2, deposit_type "hosted"
    Then the response status is 201
    And the user balance for EUR is 700.00
