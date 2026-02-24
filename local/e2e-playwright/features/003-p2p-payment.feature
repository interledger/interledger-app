Feature: Peer-to-Peer Payments
  As a verified user with KYC approved
  I want to be able to send and receive payments
  So that I can transfer funds peer-to-peer

  Background:
    Given a random test identifier is generated
    And the frontend is running at "https://interledger.test"
    And mockgatehub is running at "https://mockgatehub.interledger.test"
    And Rafiki assets are seeded

  @p2p-payment @gatehub
  Scenario: Successfully navigate to send payment page and fill payment form
    # Set up sender user with deposit
    Given the details of 'sender-user' are
      | field           | value                        |
      | emailSuffix     | sender@example.com           |
      | password        | InterlEdger2025!TestPassword |
      | country         | Germany                      |
      | firstName       | Sender                       |
      | lastName        | PayTest                      |
      | dateOfBirth     | 1988-05-15                   |
    And I complete the minimal KYC flow `sender-user`

    # Verify initial balance
    When I impersonate 'sender-user'
    And I navigate to the dashboard
    And I take a screenshot "sender-on-dashboard"

    # Navigate to send payment and fill form
    When I navigate to the send payment page
    Then I should see the payments page
    And I take a screenshot "payment-form-filled"

  @p2p-payment @gatehub
  Scenario: Successfully set up two users and navigate to payments
    Given the details of 'user-two' are
      | field           | value                        |
      | emailSuffix     | test-two@example.com         |
      | password        | InterlEdger2025!TestPassword |
      | country         | Germany                      |
      | firstName       | UserTwo                      |
      | lastName        | Test                         |
      | dateOfBirth     | 1990-01-01                   |
    And I complete the minimal KYC flow `user-two`

    When I impersonate 'user-two'
    And I navigate to the dashboard
    And I navigate to the send payment page
    Then I should see the payments page

  @p2p-payment @gatehub
  Scenario: Successfully send payment from one user to another
    # Set up sender user with KYC
    Given the details of 'sender' are
      | field           | value                        |
      | emailSuffix     | sender-p2p@example.com       |
      | password        | InterlEdger2025!TestPassword |
      | country         | Germany                      |
      | firstName       | Sender                       |
      | lastName        | User                         |
      | dateOfBirth     | 1985-06-15                   |
    And I complete the minimal KYC flow `sender`

    # Clear browser state and close context before setting up receiver
    When I clear the browser session
    And I navigate to https://interledger.test
    And I wait "2" seconds for the page to load

    # Set up receiver user with KYC in completely fresh context
    Given the details of 'receiver' are
      | field           | value                        |
      | emailSuffix     | receiver-p2p@example.com     |
      | password        | InterlEdger2025!TestPassword |
      | country         | Germany                      |
      | firstName       | Receiver                     |
      | lastName        | User                         |
      | dateOfBirth     | 1987-08-20                   |
    And I complete the minimal KYC flow `receiver`

    # Clear browser state before logging back in as sender
    When I clear the browser session
    And I navigate to https://interledger.test
    And I wait "2" seconds for the page to load
    And I impersonate 'sender'
    And I log in as myself
    And I wait "3" seconds for the page to load

    # Sender deposits funds
    And I navigate to the deposit page
    And I deposit "1000" "EUR" via the deposit iframe
    And I wait "3" seconds for the deposit to process

    # Verify sender has balance
    And I navigate to the dashboard
    And I take a screenshot "after-deposit"    
    And I should see my account balance with "1000" "EUR"
    And I take a screenshot "sender-with-balance"

    # Sender initiates payment to receiver
    And I navigate to the send payment page
    And I take a screenshot "send-payment-page-unfilled"    
    And I get the receiver wallet address for "receiver"
    And I fill in the receiver wallet address
    And I take a screenshot "send-payment-page-target"        
    And I fill in the payment amount "100"
    And I select the payment currency "EUR"
    And I take a screenshot "payment-form-ready"    
    And I submit the payment
    And I wait "5" seconds for the payment to complete

    # Verify payment succeeded
    Then I should see a payment confirmation
    And I navigate to the payments history page
    And I should see the payment in my transaction history

    When I clear the browser session
    And I navigate to https://interledger.test
    And I wait "2" seconds for the page to load
    And I impersonate 'receiver'
    And I log in as myself
    And I wait "3" seconds for the page to load
    And I should see my account balance with "100" "EUR"
    And I take a screenshot "money-received"       

  @p2p-payment @xago
  Scenario: Successfully send payment from one South African user to another
    # Set up sender user with KYC
    Given the details of 'za-sender' are
      | field           | value                        |
      | emailSuffix     | za-sender-p2p@example.com    |
      | password        | InterlEdger2025!TestPassword |
      | country         | South Africa                 |
      | firstName       | Thabo                        |
      | lastName        | Mbeki                        |
      | dateOfBirth     | 1985-06-15                   |
    And I complete the minimal KYC flow `za-sender`

    # Clear browser state and set up receiver
    When I clear the browser session
    And I navigate to https://interledger.test
    And I wait "2" seconds for the page to load

    Given the details of 'za-receiver' are
      | field           | value                        |
      | emailSuffix     | za-receiver-p2p@example.com  |
      | password        | InterlEdger2025!TestPassword |
      | country         | South Africa                 |
      | firstName       | Mandla                       |
      | lastName        | Zuma                         |
      | dateOfBirth     | 1987-08-20                   |
    And I complete the minimal KYC flow `za-receiver`

    # Log back in as sender
    When I clear the browser session
    And I navigate to https://interledger.test
    And I wait "2" seconds for the page to load
    And I impersonate 'za-sender'
    And I log in as myself
    And I wait "3" seconds for the page to load

    When I initiate a deposit for my Xago linked account
    Then my Xago specific deposit instructions should be displayed to me
    When I simulate a "800.12" "ZAR" EFT payment to Xago
    And I wait "5" seconds for the webhook to be processed
    Then I should see my balance updated to be "800.12" "ZAR"

    # Verify sender can navigate to payment page
    When I select to pay user `za-receiver`
    And continue to pay them "400.00" "ZAR" from my Xago linked account
    Then I should see my balance updated to be "400.12" "ZAR"