Feature: Plaid bank-link
  As a US wallet user
  I want to connect a bank account through Plaid
  So that I can fund my wallet from a linked bank

  # Plaid is the only US bank-link path in e2e (PLAID_ENABLED=true). The mock Plaid
  # Link overlay is an iframe served from cdn.plaid.com; the steps drive it via a
  # FrameLocator. Tartan Bank returns a stable account_id (duplicate detection);
  # Platypus Bank mints a new account_id each time (multi-account).

  Background:
    Given a random test identifier is generated
    And the frontend is running at "https://interledger.test"
    And mockplaid is running at "https://mockplaid.interledger.test"
    And Rafiki assets are seeded

  @plaid @pti
  Scenario: S1 — link a US bank via Plaid
    Given the details of 'plaid-link-user' are
      | field       | value                        |
      | emailSuffix | plaid-link@example.com       |
      | password    | InterlEdger2025!TestPassword |
      | country     | United States                |
      | firstName   | Alice                        |
      | lastName    | Smith                        |
      | dateOfBirth | 1984-06-27                   |
    And mockpti is running at "https://mockpti.interledger.test"
    And I complete the minimal PTI KYC flow `plaid-link-user`
    When I connect "Tartan Bank" "checking" via Plaid
    Then the linked account should be shown as "Plaid Checking"
    And I should have "1" Plaid bank accounts

  @plaid @pti
  Scenario: S4 — duplicate link is caught
    Given the details of 'plaid-dupe-user' are
      | field       | value                        |
      | emailSuffix | plaid-dupe@example.com       |
      | password    | InterlEdger2025!TestPassword |
      | country     | United States                |
      | firstName   | Alice                        |
      | lastName    | Smith                        |
      | dateOfBirth | 1984-06-27                   |
    And mockpti is running at "https://mockpti.interledger.test"
    And I complete the minimal PTI KYC flow `plaid-dupe-user`
    And I connect "Tartan Bank" "checking" via Plaid
    And the linked account should be shown as "Plaid Checking"
    When I connect "Tartan Bank" "checking" via Plaid
    Then I should see the snackbar "Account already linked"
    And I should have "1" Plaid bank accounts

  @plaid @pti
  Scenario: S5 — link multiple accounts from the always-new bank
    Given the details of 'plaid-multi-user' are
      | field       | value                        |
      | emailSuffix | plaid-multi@example.com      |
      | password    | InterlEdger2025!TestPassword |
      | country     | United States                |
      | firstName   | Alice                        |
      | lastName    | Smith                        |
      | dateOfBirth | 1984-06-27                   |
    And mockpti is running at "https://mockpti.interledger.test"
    And I complete the minimal PTI KYC flow `plaid-multi-user`
    And I connect "Platypus Bank" "checking" via Plaid
    And the linked account should be shown as "Plaid Checking"
    When I connect "Platypus Bank" "checking" via Plaid
    Then I should have "2" Plaid bank accounts

  @plaid @pti
  Scenario: S7 — user cancels the Plaid overlay
    Given the details of 'plaid-cancel-user' are
      | field       | value                        |
      | emailSuffix | plaid-cancel@example.com     |
      | password    | InterlEdger2025!TestPassword |
      | country     | United States                |
      | firstName   | Alice                        |
      | lastName    | Smith                        |
      | dateOfBirth | 1984-06-27                   |
    And mockpti is running at "https://mockpti.interledger.test"
    And I complete the minimal PTI KYC flow `plaid-cancel-user`
    And I navigate to "/connect/bank"
    When I cancel the Plaid overlay
    Then I should be navigated to dashboard "home"

  @plaid @gatehub
  Scenario: S9 — non-US user is gated out of the Plaid page
    Given the details of 'plaid-gated-user' are
      | field       | value                        |
      | emailSuffix | plaid-gated@example.com      |
      | password    | InterlEdger2025!TestPassword |
      | country     | Germany                      |
      | firstName   | Alice                        |
      | lastName    | Smith                        |
      | dateOfBirth | 1984-06-27                   |
    And mockgatehub is running at "https://mockgatehub.interledger.test"
    And I complete the minimal KYC flow `plaid-gated-user`
    When I navigate to "/connect/bank"
    Then I should be navigated to dashboard "home"

  @plaid @pti
  Scenario: S12 — link failure surfaces the error
    Given the details of 'plaid-fail-user' are
      | field       | value                        |
      | emailSuffix | plaid-fail@example.com       |
      | password    | InterlEdger2025!TestPassword |
      | country     | United States                |
      | firstName   | Alice                        |
      | lastName    | Smith                        |
      | dateOfBirth | 1984-06-27                   |
    And mockpti is running at "https://mockpti.interledger.test"
    And I complete the minimal PTI KYC flow `plaid-fail-user`
    When I connect "Failing Bank" "checking" via Plaid
    Then I should see the snackbar "Bank account linking failed. Please try again."
    And I should have "0" Plaid bank accounts

  @plaid @deposit @pti
  Scenario: S11 — a Plaid-linked account funds a PTI deposit
    Given the details of 'plaid-deposit-user' are
      | field       | value                        |
      | emailSuffix | plaid-deposit@example.com    |
      | password    | InterlEdger2025!TestPassword |
      | country     | United States                |
      | firstName   | Alice                        |
      | lastName    | Smith                        |
      | dateOfBirth | 1984-06-27                   |
    And mockpti is running at "https://mockpti.interledger.test"
    And I complete the minimal PTI KYC flow `plaid-deposit-user`
    And I connect "Tartan Bank" "checking" via Plaid
    When I navigate to the deposit page
    And I deposit "100" "USD" via the PTI deposit form
    Then I should see my balance updated with "100" "USD"
