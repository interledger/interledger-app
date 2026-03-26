Feature: Deposit Funds
  As a verified user with KYC approved
  I want to deposit EUR into my wallet
  So that I can have funds available for payments

  Background:
    Given a random test identifier is generated
    And the frontend is running at "https://interledger.test"
    And mockgatehub is running at "https://mockgatehub.interledger.test"
    And mockxago is running at "https://mockxago.interledger.test"
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

  @deposit @pti @return
  Scenario: Deposit is returned after settling and balance reverts to initial
    Given the details of 'deposit-return-pti-user' are
      | field           | value                          |
      | emailSuffix     | alice-return-pti@example.com   |
      | password        | InterlEdger2025!TestPassword   |
      | country         | United States                  |
      | firstName       | Alice                          |
      | lastName        | Smith                          |
      | dateOfBirth     | 1984-06-27                     |
    And mockpti is running at "https://mockpti.interledger.test"
    And I complete the minimal PTI KYC flow `deposit-return-pti-user`
    When I connect a US bank account
    And I navigate to the deposit page
    And I deposit "100" "USD" via the PTI deposit form
    Then I should see my balance updated with "100" "USD"
    When mockpti returns the deposit
    Then I should see my balance updated with "0" "USD"

  @deposit @pti
  Scenario: Successfully deposit 100 USD into wallet as a USA user
    Given the details of 'deposit-pti-user' are
      | field           | value                        |
      | emailSuffix     | alice-pti@example.com        |
      | password        | InterlEdger2025!TestPassword |
      | country         | United States                |
      | firstName       | Alice                        |
      | lastName        | Smith                        |
      | dateOfBirth     | 1984-06-27                   |
    And mockpti is running at "https://mockpti.interledger.test"
    And I complete the minimal PTI KYC flow `deposit-pti-user`
    When I connect a US bank account
    And I navigate to the deposit page
    And I deposit "100" "USD" via the PTI deposit form
    Then I should see my balance updated with "100" "USD"