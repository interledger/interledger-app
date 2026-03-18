Feature: Deposit Funds
  As a verified user with KYC approved
  I want to deposit funds into my wallet
  So that I can have funds available for payments

  @deposit @gatehub
  Scenario: Successfully deposit 100 EUR into wallet via GateHub iframe
    Given a random test identifier is generated
    And the frontend is running at "https://interledger.test"
    And mockgatehub is running at "https://mockgatehub.interledger.test"
    And Rafiki assets are seeded
    Given the details of 'deposit-user' are
      | field           | value                        |
      | emailSuffix     | alice@example.com            |
      | password        | InterlEdger2025!TestPassword |
      | country         | Germany                      |
      | firstName       | Alice                        |
      | lastName        | Smith                        |
      | dateOfBirth     | 1984-06-27                   |
    And I complete the minimal KYC flow `deposit-user`
    When I navigate to the deposit page
    And I deposit "100" "EUR" via the deposit iframe
    Then I should see my balance updated with "100" "EUR"

  @deposit @fees @gatehub
  Scenario: Successfully deposit 100 EUR into wallet with 1% fee
    Given a random test identifier is generated
    And the frontend is running at "https://interledger.test"
    And mockgatehub is running at "https://mockgatehub.interledger.test"
    And Rafiki assets are seeded
    Given that Gatehub charges my user a 1% deposit fee
    And the details of 'deposit-fee-user' are
      | field           | value                        |
      | emailSuffix     | alice-fee@example.com        |
      | password        | InterlEdger2025!TestPassword |
      | country         | Germany                      |
      | firstName       | Alice                        |
      | lastName        | Fee                          |
      | dateOfBirth     | 1984-06-27                   |
    And I complete the minimal KYC flow `deposit-fee-user`
    When I navigate to the deposit page
    And I deposit "100" "EUR" via the deposit iframe
    Then I should see my balance updated with "99" "EUR"

  @deposit @xago
  Scenario: Successfully deposit 500 ZAR into wallet via MockXago test deposit
    Given a random test identifier is generated
    And the frontend is running at "https://interledger.test"
    And mockxago is running at "https://mockxago.interledger.test"
    And Rafiki assets are seeded
    And the details of 'xago-deposit-user' are
      | field           | value                        |
      | emailSuffix     | alice-za@example.com         |
      | password        | InterlEdger2025!TestPassword |
      | country         | South Africa                 |
      | firstName       | Alice                        |
      | lastName        | Nkosi                        |
      | dateOfBirth     | 1984-06-27                   |
    And I complete the minimal KYC flow `xago-deposit-user`
    When I initiate a deposit for my Xago linked account
    Then my Xago specific deposit instructions should be displayed to me
    When I simulate a "500" "ZAR" EFT payment to Xago
    And I wait "5" seconds for the webhook to be processed
    Then I should see my balance updated with "500" "ZAR"
