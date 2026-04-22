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
  Scenario: Successfully deposit and withdraw 50 EUR from Germany based wallet
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

  @withdrawal @pti @return
  Scenario: Balance goes negative when deposit is returned after withdrawal from PTI wallet
    Given the details of 'withdrawal-return-pti-user' are
      | field           | value                        |
      | emailSuffix     | bob-return-pti@example.com   |
      | password        | InterlEdger2025!TestPassword |
      | country         | United States                |
      | firstName       | Bob                          |
      | lastName        | Johnson                      |
      | dateOfBirth     | 1990-03-15                   |
    And mockpti is running at "https://mockpti.interledger.test"
    And I complete the minimal PTI KYC flow `withdrawal-return-pti-user`
    When I connect a US bank account
    And I navigate to the deposit page
    And I deposit "100" "USD" via the PTI deposit form
    Then I should see my balance updated with "100" "USD"

    When I navigate to the withdrawal page
    And I withdraw "50" "USD" via the PTI withdrawal form
    Then I should see my balance updated with "50" "USD"

    When mockpti returns the deposit
    Then I should see my balance updated with "-50" "USD"

  @withdrawal @pti
  Scenario: Successfully deposit and withdraw 50 USD as a USA user
    Given the details of 'withdrawal-pti-user' are
      | field           | value                        |
      | emailSuffix     | bob-pti@example.com          |
      | password        | InterlEdger2025!TestPassword |
      | country         | United States                |
      | firstName       | Bob                          |
      | lastName        | Johnson                      |
      | dateOfBirth     | 1990-03-15                   |
    And mockpti is running at "https://mockpti.interledger.test"
    And I complete the minimal PTI KYC flow `withdrawal-pti-user`
    When I connect a US bank account
    And I navigate to the deposit page
    And I deposit "100" "USD" via the PTI deposit form
    Then I should see my balance updated with "100" "USD"

    When I navigate to the withdrawal page
    And I withdraw "50" "USD" via the PTI withdrawal form
    Then I should see my balance updated with "50" "USD"
