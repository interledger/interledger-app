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
