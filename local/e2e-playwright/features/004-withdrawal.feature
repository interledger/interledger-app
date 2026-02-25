Feature: Withdraw Funds
  As a verified user with KYC approved and deposited funds
  I want to withdraw EUR from my wallet
  So that I can transfer funds to my external bank account

  Background:
    Given a random test identifier is generated
    And the frontend is running at "https://interledger.test"
    And mockgatehub is running at "https://mockgatehub.interledger.test"
    And mockxago is running at "https://mockxago.interledger.test"
    And Rafiki assets are seeded

  @withdrawal @gatehub
  Scenario: Successfully deposit and withdraw 50 EUR from wallet
    Given the details of 'withdrawal-user' are
      | field           | value                        |
      | emailSuffix     | bob@example.com              |
      | password        | InterlEdger2025!TestPassword |
      | country         | Germany                      |
      | firstName       | Bob                          |
      | lastName        | Johnson                      |
      | dateOfBirth     | 1990-03-15                   |
    And I complete the minimal KYC flow `withdrawal-user`
    When I navigate to the deposit page
    And I deposit "100" "EUR" via the deposit iframe
    Then I should see my balance updated with "100" "EUR"
    
    When I navigate to the withdrawal page
    And I withdraw "50" "EUR" via the withdrawal iframe
    Then I should see my balance updated with "50" "EUR"

  @withdrawal @gatehub @fees
  Scenario: Successfully deposit 100 EUR and withdraw with 2% fee
    Given the details of 'withdrawal-user' are
      | field           | value                        |
      | emailSuffix     | bob@example.com              |
      | password        | InterlEdger2025!TestPassword |
      | country         | Germany                      |
      | firstName       | Bob                          |
      | lastName        | Johnson                      |
      | dateOfBirth     | 1990-03-15                   |
    And I complete the minimal KYC flow `withdrawal-user`
    Given that Gatehub charges my user a 2% withdrawal fee
    When I navigate to the deposit page
    And I deposit "100" "EUR" via the deposit iframe
    Then I should see my balance updated with "100" "EUR"
    
    When I navigate to the withdrawal page
    And I withdraw "50" "EUR" via the withdrawal iframe
    Then I should see my balance updated with "49" "EUR"

  @withdrawal @xago
  Scenario: South African user withdraws to a linked account
    Given the details of 'xago-withdrawal-user' are
      | field           | value                        |
      | emailSuffix     | bob-za@example.com           |
      | password        | InterlEdger2025!TestPassword |
      | country         | South Africa                 |
      | firstName       | Bob                          |
      | lastName        | Zuma                         |
      | dateOfBirth     | 1990-03-15                   |
    And I complete the minimal KYC flow `xago-withdrawal-user`
    And I linked a SA bank account with "First National Bank" and account number "6208889997"
    And I deposited "2000.00" "ZAR" into my xago backed wallet

    When I navigate to the withdrawal page

    Then I should see the "Withdraw" form
    And I should see text "You have R 2000.00 available in your balance." on the page
    And I should see text "0.00" on the page

    When I set the withdraw amount to "99.99"
    And I select the first available linked account to withdraw to
    And I set the withdraw note to "withdraw note text"
    And I press on "Continue"

    Then I should see the "Confirm withdraw" form
    When I press on "Confirm withdraw"
    And I wait "4" seconds for the page to load

    Then I should be navigated to dashboard "Home"
    And I should see my balance updated with "1900.01" "ZAR"
