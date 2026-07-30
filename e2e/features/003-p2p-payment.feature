Feature: Peer-to-Peer Payments
  As a verified user with KYC approved
  I want to be able to send and receive payments
  So that I can transfer funds peer-to-peer

  Background:
    Given a random test identifier is generated
    And the frontend is running at "https://interledger.test"
    And mockgatehub is running at "https://mockgatehub.interledger.test"
    And mockxago is running at "https://mockxago.interledger.test"
    And Rafiki assets are seeded

  @p2p-payment @gatehub
  Scenario: Successfully navigate to send payment page and fill payment form for Germany based user
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
  Scenario: Successfully set up two Germany based users and navigate to payments
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
  Scenario: Successfully send payment from one Germany based user to another Germany based user
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
    # Double-submit regression: one payment, redirect Home, not stranded on Pay search.
    And I rapidly double-click to submit the payment
    And I wait for the payment to complete

    # Verify payment succeeded and post-confirm navigation is healthy
    Then I should see a payment confirmation
    And I should be redirected to the home page and able to navigate away
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
    Then I should see my balance updated with "800.12" "ZAR"

    # Sender initiates payment to receiver
    When I navigate to the send payment page
    And I get the receiver wallet address for "za-receiver"
    And I fill in the receiver wallet address
    And I fill in the payment amount "400.00"
    And I select the payment currency "ZAR"
    And I submit the payment
    And I wait for the payment to complete
    Then I should see a payment confirmation
    And I should see my balance updated with "400.12" "ZAR"

    # Verify receiver sees the funds
    When I clear the browser session
    And I navigate to https://interledger.test
    And I wait "2" seconds for the page to load
    And I impersonate 'za-receiver'
    And I log in as myself
    And I wait "3" seconds for the page to load
    Then I should see my balance updated with "400.00" "ZAR"
    And I take a screenshot "za-receiver-balance"

  @p2p-payment @cross-provider @xago @gatehub
  Scenario: Successfully send a cross-provider payment from a South African user to a Germany based user
    # Set up ZAR/Xago sender with KYC
    Given the details of 'cp-za-de-sender' are
      | field           | value                          |
      | emailSuffix     | cp-za-de-sender@example.com    |
      | password        | InterlEdger2025!TestPassword   |
      | country         | South Africa                   |
      | firstName       | ZASender                       |
      | lastName        | CP                             |
      | dateOfBirth     | 1986-03-11                     |
    And I complete the minimal KYC flow `cp-za-de-sender`
    And I enable cross-provider payments for "cp-za-de-sender"

    # Clear browser state and set up EUR/Gatehub receiver
    When I clear the browser session
    And I navigate to https://interledger.test
    And I wait "2" seconds for the page to load

    Given the details of 'cp-za-de-receiver' are
      | field           | value                           |
      | emailSuffix     | cp-za-de-receiver@example.com   |
      | password        | InterlEdger2025!TestPassword    |
      | country         | Germany                         |
      | firstName       | DEReceiver                      |
      | lastName        | CP                              |
      | dateOfBirth     | 1991-09-23                      |
    And I complete the minimal KYC flow `cp-za-de-receiver`
    And I enable cross-provider payments for "cp-za-de-receiver"

    # Log back in as the ZAR/Xago sender
    When I clear the browser session
    And I navigate to https://interledger.test
    And I wait "2" seconds for the page to load
    And I impersonate 'cp-za-de-sender'
    And I log in as myself
    And I wait "3" seconds for the page to load

    When I initiate a deposit for my Xago linked account
    Then my Xago specific deposit instructions should be displayed to me
    When I simulate a "5500.00" "ZAR" EFT payment to Xago
    And I wait "5" seconds for the webhook to be processed
    Then I should see my balance updated with "5500.00" "ZAR"

    # Sender pays the EUR/Gatehub receiver from their ZAR balance
    When I navigate to the send payment page
    And I get the receiver wallet address for "cp-za-de-receiver"
    And I fill in the receiver wallet address
    And I fill in the payment amount "5000.00"
    And I select the payment currency "ZAR"
    And I take a screenshot "cross-provider-za-to-de-payment-ready"
    And I submit the payment
    And I wait for the payment to complete
    Then I should see a payment confirmation
    And I should see my balance updated with "500.00" "ZAR"

    # Verify the Gatehub receiver sees the converted EUR funds
    # ZAR -> EUR mock rate is 0.051218, so 5000.00 ZAR converts to 256.09 EUR
    When I clear the browser session
    And I navigate to https://interledger.test
    And I wait "2" seconds for the page to load
    And I impersonate 'cp-za-de-receiver'
    And I log in as myself
    And I wait "3" seconds for the page to load
    Then I should see my balance updated with "256.09" "EUR"
    And I take a screenshot "cross-provider-za-to-de-receiver-balance"

  @p2p-payment @cross-provider @gatehub @xago
  Scenario: Successfully send a cross-provider payment from a Germany based user to a South African user
    # Set up EUR/Gatehub sender with KYC
    Given the details of 'cp-de-za-sender' are
      | field           | value                        |
      | emailSuffix     | cp-de-za-sender@example.com  |
      | password        | InterlEdger2025!TestPassword |
      | country         | Germany                      |
      | firstName       | DESender                     |
      | lastName        | CP                           |
      | dateOfBirth     | 1983-12-02                   |
    And I complete the minimal KYC flow `cp-de-za-sender`
    And I enable cross-provider payments for "cp-de-za-sender"

    # Clear browser state and set up ZAR/Xago receiver
    When I clear the browser session
    And I navigate to https://interledger.test
    And I wait "2" seconds for the page to load

    Given the details of 'cp-de-za-receiver' are
      | field           | value                           |
      | emailSuffix     | cp-de-za-receiver@example.com   |
      | password        | InterlEdger2025!TestPassword    |
      | country         | South Africa                    |
      | firstName       | ZAReceiver                      |
      | lastName        | CP                              |
      | dateOfBirth     | 1989-07-19                      |
    And I complete the minimal KYC flow `cp-de-za-receiver`
    And I enable cross-provider payments for "cp-de-za-receiver"

    # Log back in as the EUR/Gatehub sender
    When I clear the browser session
    And I navigate to https://interledger.test
    And I wait "2" seconds for the page to load
    And I impersonate 'cp-de-za-sender'
    And I log in as myself
    And I wait "3" seconds for the page to load

    And I navigate to the deposit page
    And I deposit "1000" "EUR" via the deposit iframe
    And I wait "3" seconds for the deposit to process
    And I navigate to the dashboard
    And I should see my account balance with "1000" "EUR"

    # Sender pays the ZAR/Xago receiver from their EUR balance
    And I navigate to the send payment page
    And I get the receiver wallet address for "cp-de-za-receiver"
    And I fill in the receiver wallet address
    And I fill in the payment amount "200.00"
    And I select the payment currency "EUR"
    And I take a screenshot "cross-provider-de-to-za-payment-ready"
    And I submit the payment
    And I wait for the payment to complete
    Then I should see a payment confirmation
    And I should see my balance updated with "800.00" "EUR"

    # Verify the Xago receiver sees the converted ZAR funds
    # EUR -> ZAR mock rate is 19.524, so 200.00 EUR converts to 3904.80 ZAR
    When I clear the browser session
    And I navigate to https://interledger.test
    And I wait "2" seconds for the page to load
    And I impersonate 'cp-de-za-receiver'
    And I log in as myself
    And I wait "3" seconds for the page to load
    Then I should see my balance updated with "3904.80" "ZAR"
    And I take a screenshot "cross-provider-de-to-za-receiver-balance"
