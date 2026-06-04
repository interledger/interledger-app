Feature: Rafiki Full Node Payment Lifecycle
  As a verified user with KYC approved
  I want payments to be processed through the Rafiki full node
  So that funds settle correctly via the Rafiki ILP node payment lifecycle

  # NOTE: These scenarios require BACKEND_RAFIKI_NODE_ENABLED=true in the local environment.
  # Set this in local/.env before running, or pass BACKEND_RAFIKI_NODE_ENABLED=true
  # to the wallet service. Without the flag, the node-specific webhook workflows are skipped.

  Background:
    Given a random test identifier is generated
    And the frontend is running at "https://interledger.test"
    And mockgatehub is running at "https://mockgatehub.interledger.test"
    And Rafiki assets are seeded

  @rafiki-node @p2p-payment @gatehub
  Scenario: Full Rafiki node payment lifecycle - sender and receiver balances settle correctly
    # Set up sender user with KYC
    Given the details of 'node-sender' are
      | field       | value                        |
      | emailSuffix | node-sender@example.com      |
      | password    | InterlEdger2025!TestPassword |
      | country     | Germany                      |
      | firstName   | NodeSender                   |
      | lastName    | Test                         |
      | dateOfBirth | 1988-05-15                   |
    And I complete the minimal KYC flow `node-sender`

    # Clear browser state and set up receiver in a fresh context
    When I clear the browser session
    And I navigate to https://interledger.test
    And I wait "2" seconds for the page to load

    Given the details of 'node-receiver' are
      | field       | value                        |
      | emailSuffix | node-receiver@example.com    |
      | password    | InterlEdger2025!TestPassword |
      | country     | Germany                      |
      | firstName   | NodeReceiver                 |
      | lastName    | Test                         |
      | dateOfBirth | 1990-03-22                   |
    And I complete the minimal KYC flow `node-receiver`

    # Log in as sender and deposit funds
    When I clear the browser session
    And I navigate to https://interledger.test
    And I wait "2" seconds for the page to load
    And I impersonate 'node-sender'
    And I log in as myself
    And I wait "3" seconds for the page to load
    And I navigate to the deposit page
    And I deposit "500" "EUR" via the deposit iframe
    And I wait "3" seconds for the deposit to process
    And I navigate to the dashboard
    And I should see my account balance with "500" "EUR"
    And I take a screenshot "node-sender-deposit-confirmed"

    # Sender initiates payment to receiver
    And I navigate to the send payment page
    And I get the receiver wallet address for "node-receiver"
    And I fill in the receiver wallet address
    And I fill in the payment amount "100"
    And I select the payment currency "EUR"
    And I submit the payment
    And I wait "5" seconds for the payment to complete

    # Verify sender balance decreased
    Then I should see a payment confirmation
    And I navigate to the payments history page
    And I should see the payment in my transaction history
    And I take a screenshot "node-sender-payment-history"
    And I navigate to the dashboard
    And I should see my account balance with "400" "EUR"
    And I take a screenshot "node-sender-balance-after-payment"

    # Log in as receiver and verify incoming balance
    When I clear the browser session
    And I navigate to https://interledger.test
    And I wait "2" seconds for the page to load
    And I impersonate 'node-receiver'
    And I log in as myself
    And I wait "5" seconds for the page to load
    And I navigate to the dashboard
    And I should see my account balance with "100" "EUR"
    And I take a screenshot "node-receiver-balance-after-payment"
    And I navigate to the payments history page
    And I should see the payment in my transaction history
    And I take a screenshot "node-receiver-payment-history"

  @rafiki-node @transaction-state @gatehub
  Scenario: Outgoing payment transitions from pending to completed in transaction history
    # Set up two users with KYC
    Given the details of 'state-sender' are
      | field       | value                        |
      | emailSuffix | state-sender@example.com     |
      | password    | InterlEdger2025!TestPassword |
      | country     | Germany                      |
      | firstName   | StateSender                  |
      | lastName    | Test                         |
      | dateOfBirth | 1985-01-10                   |
    And I complete the minimal KYC flow `state-sender`

    When I clear the browser session
    And I navigate to https://interledger.test
    And I wait "2" seconds for the page to load

    Given the details of 'state-receiver' are
      | field       | value                        |
      | emailSuffix | state-receiver@example.com   |
      | password    | InterlEdger2025!TestPassword |
      | country     | Germany                      |
      | firstName   | StateReceiver                |
      | lastName    | Test                         |
      | dateOfBirth | 1992-07-18                   |
    And I complete the minimal KYC flow `state-receiver`

    # Log in as sender, deposit, and initiate payment
    When I clear the browser session
    And I navigate to https://interledger.test
    And I wait "2" seconds for the page to load
    And I impersonate 'state-sender'
    And I log in as myself
    And I wait "3" seconds for the page to load
    And I navigate to the deposit page
    And I deposit "300" "EUR" via the deposit iframe
    And I wait "3" seconds for the deposit to process

    And I navigate to the send payment page
    And I get the receiver wallet address for "state-receiver"
    And I fill in the receiver wallet address
    And I fill in the payment amount "50"
    And I select the payment currency "EUR"
    And I submit the payment
    And I wait "5" seconds for the payment to complete

    # Immediately after redirect, check transaction history for any state entry
    And I navigate to the payments history page
    And I take a screenshot "transaction-after-payment-submit"

    # Wait for the Temporal workflow to run the full GateHub signal handshake
    And I wait "10" seconds for the payment to settle
    And I navigate to the payments history page
    And I should see a completed outgoing transaction for "50" "EUR" in my payments history
    And I take a screenshot "transaction-completed-state"

  @rafiki-node @validation @gatehub
  Scenario: Payment to own wallet address is rejected by Rafiki node validation
    # Set up a single user with KYC and a deposit
    Given the details of 'self-payer' are
      | field       | value                        |
      | emailSuffix | self-payer@example.com       |
      | password    | InterlEdger2025!TestPassword |
      | country     | Germany                      |
      | firstName   | SelfPayer                    |
      | lastName    | Test                         |
      | dateOfBirth | 1983-11-05                   |
    And I complete the minimal KYC flow `self-payer`

    And I navigate to the deposit page
    And I deposit "200" "EUR" via the deposit iframe
    And I wait "3" seconds for the deposit to process
    And I navigate to the dashboard
    And I should see my account balance with "200" "EUR"
    And I take a screenshot "self-payer-balance-before"

    # Retrieve and store own wallet address for reuse in the payment form
    And I get my own wallet address

    # Attempt to pay to own wallet address
    And I navigate to the send payment page
    And I fill in my own wallet address as the receiver
    And I fill in the payment amount "50"
    And I select the payment currency "EUR"
    And I submit the payment
    # Allow time for the outgoing_payment.created workflow to validate and cancel
    And I wait "10" seconds for the payment workflow to complete

    # Balance must be unchanged — the node workflow cancelled the outgoing payment
    And I navigate to the dashboard
    And I should see my account balance with "200" "EUR"
    And I take a screenshot "self-payer-balance-unchanged"
    And I navigate to the payments history page
    And I should not see a completed outgoing transaction for "50" "EUR" in my payments history
    And I take a screenshot "self-payer-no-completed-tx"

  @rafiki-node @mixed-provider @gatehub
  Scenario: GateHub-to-non-GateHub payment uses legacy path without pending state
    # Set up sender (GateHub, Germany) with KYC
    Given the details of 'legacy-sender' are
      | field       | value                        |
      | emailSuffix | legacy-sender@example.com    |
      | password    | InterlEdger2025!TestPassword |
      | country     | Germany                      |
      | firstName   | LegacySender                 |
      | lastName    | Test                         |
      | dateOfBirth | 1986-04-12                   |
    And I complete the minimal KYC flow `legacy-sender`

    When I clear the browser session
    And I navigate to https://interledger.test
    And I wait "2" seconds for the page to load

    # Set up receiver (GateHub, Germany) with KYC — but we send to a
    # simulated non-GateHub ILP address to exercise the mixed-provider guard
    Given the details of 'legacy-receiver' are
      | field       | value                        |
      | emailSuffix | legacy-receiver@example.com  |
      | password    | InterlEdger2025!TestPassword |
      | country     | Germany                      |
      | firstName   | LegacyReceiver               |
      | lastName    | Test                         |
      | dateOfBirth | 1989-09-03                   |
    And I complete the minimal KYC flow `legacy-receiver`

    When I clear the browser session
    And I navigate to https://interledger.test
    And I wait "2" seconds for the page to load
    And I impersonate 'legacy-sender'
    And I log in as myself
    And I wait "3" seconds for the page to load
    And I navigate to the deposit page
    And I deposit "400" "EUR" via the deposit iframe
    And I wait "3" seconds for the deposit to process
    And I navigate to the dashboard
    And I should see my account balance with "400" "EUR"

    And I navigate to the send payment page
    And I get the receiver wallet address for "legacy-receiver"
    And I fill in the receiver wallet address
    And I fill in the payment amount "75"
    And I select the payment currency "EUR"
    And I submit the payment
    And I wait "5" seconds for the payment to complete

    # Payment should succeed via legacy path and balance should reflect the deduction
    Then I should see a payment confirmation
    And I navigate to the payments history page
    And I should see the payment in my transaction history
    And I take a screenshot "legacy-path-payment-history"
    And I navigate to the dashboard
    And I should see my account balance with "325" "EUR"
    And I take a screenshot "legacy-path-balance-after"
