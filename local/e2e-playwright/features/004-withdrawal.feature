Feature: Withdraw Funds
  As a verified user with KYC approved and deposited funds
  I want to withdraw EUR from my wallet
  So that I can transfer funds to my external bank account

  Background:
    Given a random test identifier is generated
    And the frontend is running at "https://interledger.test"
    And mockgatehub is running at "https://mockgatehub.interledger.test"
    And Rafiki assets are seeded
    Given the details of 'withdrawal-user' are
      | field           | value                        |
      | emailSuffix     | bob@example.com              |
      | password        | InterlEdger2025!TestPassword |
      | country         | Germany                      |
      | firstName       | Bob                          |
      | lastName        | Johnson                      |
      | dateOfBirth     | 1990-03-15                   |
    And I complete the minimal KYC flow `withdrawal-user`

  @withdrawal @gatehub
  Scenario: Successfully deposit and withdraw 50 EUR from wallet
    When I navigate to the deposit page
    And I deposit "100" "EUR" via the deposit iframe
    Then I should see my balance updated with "100" "EUR"
    
    When I navigate to the withdrawal page
    And I withdraw "50" "EUR" via the withdrawal iframe
    Then I should see my balance updated with "50" "EUR"

  @withdrawal @gatehub @fees
  Scenario: Successfully deposit 100 EUR and withdraw with 2% fee
    Given that Gatehub charges my user a 2% withdrawal fee
    When I navigate to the deposit page
    And I deposit "100" "EUR" via the deposit iframe
    Then I should see my balance updated with "100" "EUR"
    
    When I navigate to the withdrawal page
    And I withdraw "50" "EUR" via the withdrawal iframe
    Then I should see my balance updated with "49" "EUR"

  @withdrawal @xago
  Scenario: Withdrawal page not available for South African users with linked accounts
    Given the details of 'xago-withdrawal-user' are
      | field           | value                        |
      | emailSuffix     | bob-za@example.com           |
      | password        | InterlEdger2025!TestPassword |
      | country         | South Africa                 |
      | firstName       | Bob                          |
      | lastName        | Zuma                         |
      | dateOfBirth     | 1990-03-15                   |
    And I complete the minimal KYC flow `xago-withdrawal-user`
    When I navigate to the withdrawal page
    And I wait "3" seconds for the page to load
    Then I should see text "404" on the page
    And I should see text "An error occurred" on the page
    # Note: Like deposits, withdrawal page is not available for Xago users with
    # auto-created linked accounts. Withdrawals for Xago users likely use a
    # different interface/flow specific to South African banking requirements.
