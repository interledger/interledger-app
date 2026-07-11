Feature: Botanist Wallet Settings
  As an admin
  I want to manage a wallet's settings from the Wallet Settings tab
  So that I can change entityconf-backed wallet confs without a deploy

  Background:
    Given a random test identifier is generated
    And the frontend is running at "https://interledger.test"
    And the admin portal is running at "https://admin.mgnt.interledger.test"
    And mockgatehub is running at "https://mockgatehub.interledger.test"
    Given the details of 'botanist-user' are
      | field       | value                          |
      | emailSuffix | botanist-wallet-settings@example.com |
      | password    | InterlEdger2025!TestPassword   |
      | firstName   | Botanist                       |
      | lastName    | Settings                       |
      | dateOfBirth | 1990-06-15                     |
    And I impersonate 'botanist-user'

  @botanist @wallet-settings-toggle @gatehub
  Scenario: Admin can toggle a wallet setting and it persists after navigating away and back
    Given that my "country" is "germany"
    And I complete the minimal KYC flow `botanist-user`
    When I navigate to the admin portal
    And I navigate to my wallet settings page in the admin portal
    And I take a screenshot "wallet-settings-loaded"
    Then the "Send" wallet setting toggle should be off
    When I toggle the "Send" wallet setting on
    Then the "Send" wallet setting should be enabled in the database for my wallet
    And I take a screenshot "wallet-setting-toggled-on"
    When I navigate to my wallet profile page in the admin portal
    And I navigate to my wallet settings page in the admin portal
    Then the "Send" wallet setting toggle should be on
    And I take a screenshot "wallet-setting-persisted-after-navigation"
