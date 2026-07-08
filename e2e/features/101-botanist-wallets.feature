Feature: Botanist Wallets Filter
  As an admin
  I want to filter the wallets list by common fields
  So that I can quickly locate a specific wallet

  Background:
    Given a random test identifier is generated
    And the frontend is running at "https://interledger.test"
    And the admin portal is running at "https://admin.mgnt.interledger.test"
    And mockgatehub is running at "https://mockgatehub.interledger.test"
    And mockxago is running at "https://mockxago.interledger.test"
    Given the details of 'botanist-user' are
      | field       | value                        |
      | emailSuffix | botanist@example.com         |
      | password    | InterlEdger2025!TestPassword |
      | firstName   | Botanist                     |
      | lastName    | Tester                       |
      | dateOfBirth | 1990-06-15                   |
    And I impersonate 'botanist-user'

  @botanist @wallets-filter @xago
  Scenario: Admin can search wallets by wallet name after a user signs up
    Given that my "country" is "South Africa"
    And I completed the signup workflow
    And I take a screenshot "signup-complete"
    And I completed the account verification workflow
    And I take a screenshot "account-verified"
    And I finished the phone confirmation workflow
    And I take a screenshot "phone-verified"
    And I finished the TOTP registration workflow
    And I take a screenshot "totp-registered"
    And I finished the wallet address creation workflow
    And I take a screenshot "wallet-address-created"
    When I navigate to the admin portal
    And I take a screenshot "admin-portal-loaded"
    And I navigate to the botanist wallets page
    And I take a screenshot "wallets-page-unfiltered"
    Then my wallet should appear in the wallets list
    And the wallets list should have more than 1 result
    And I take a screenshot "wallet-visible-unfiltered"
    When I filter the wallets list by my wallet name
    And I take a screenshot "filter-applied"
    Then my wallet should appear in the wallets list
    And the wallets list should show exactly 1 result
    And I take a screenshot "wallet-visible-filtered"

  @botanist @wallets-totp-reset @xago
  Scenario: Admin can reset a wallet user's authenticator enrollment
    Given that my "country" is "South Africa"
    And I completed the signup workflow
    And I completed the account verification workflow
    And I finished the phone confirmation workflow
    And I take a screenshot "phone-verified"
    And I finished the TOTP registration workflow
    And I finished the wallet address creation workflow
    When I navigate to the admin portal
    And I navigate to my wallet profile page in the admin portal
    Then the reset authenticator button should be visible
    When I click the reset authenticator button
    Then the authenticator reset confirmation modal should be visible
    And I take a screenshot "reset-authenticator-modal"
    When I confirm the authenticator reset
    Then the reset authenticator button should not be visible
    And my TOTP should be disabled
    And an authenticator reset audit log entry should exist
    And I take a screenshot "reset-authenticator-complete"
    When I start a new browser session
    And I navigate to the login page
    And I fill in the login form with my details
    And I submit the login
    Then I should be navigated to the TOTP page
    And I take a screenshot "totp-reenrollment-after-admin-reset"

  @botanist @wallets-features-toggle @gatehub
  Scenario: Admin can toggle the deleteAccountEnabled feature for a wallet
    Given that my "country" is "germany"
    And I complete the minimal KYC flow `botanist-user`
    When I navigate to the admin portal
    And I navigate to my wallet profile page in the admin portal
    Then the "deleteAccountEnabled" feature toggle should be off
    When I toggle the "deleteAccountEnabled" feature on
    Then the "deleteAccountEnabled" feature should be enabled in the database for my wallet
    And I take a screenshot "delete-account-feature-toggled-on"
