Feature: Deposit Funds
  As a verified user with KYC approved
  I want to deposit EUR into my wallet
  So that I can have funds available for payments

  Background:
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

  @deposit @gatehub
  Scenario: Successfully deposit 100 EUR into wallet
    When I navigate to the deposit page
    And I deposit "100" "EUR" via the deposit iframe
    Then I should see my balance updated with "100" "EUR"

  @deposit @fees @gatehub
  Scenario: Successfully deposit 100 EUR into wallet with 1% fee
    Given that Gatehub charges my user a 1% deposit fee
    When I navigate to the deposit page
    And I deposit "100" "EUR" via the deposit iframe
    Then I should see my balance updated with "99" "EUR"