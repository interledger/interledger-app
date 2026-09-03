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
  Scenario: Admin can search wallets by each individual filter field, and by all of them combined
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
    And I navigate to the personal details page to activate wallet
    And I wait for the KYC iframe to load
    And I fill and submit the mockxago KYC iframe
    And I wait for the KYC completion
    And I take a screenshot "kyc-completed"
    When I navigate to the admin portal
    And I take a screenshot "admin-portal-loaded"
    And I navigate to the botanist wallets page
    And I take a screenshot "wallets-page-unfiltered"
    Then my wallet should appear in the wallets list
    And the wallets list should have more than 1 result
    And I take a screenshot "wallet-visible-unfiltered"

    When I filter the wallets list by my first name
    And I take a screenshot "filter-applied-first-name"
    Then my wallet should appear in the wallets list
    And the wallets list should show exactly 1 result
    And I take a screenshot "wallet-visible-filtered-first-name"

    When I navigate to the botanist wallets page
    And I filter the wallets list by my last name
    And I take a screenshot "filter-applied-last-name"
    Then my wallet should appear in the wallets list
    And the wallets list should show exactly 1 result
    And I take a screenshot "wallet-visible-filtered-last-name"

    When I navigate to the botanist wallets page
    And I filter the wallets list by my email
    And I take a screenshot "filter-applied-email"
    Then my wallet should appear in the wallets list
    And the wallets list should show exactly 1 result
    And I take a screenshot "wallet-visible-filtered-email"

    When I navigate to the botanist wallets page
    And I filter the wallets list by my phone number
    And I take a screenshot "filter-applied-phone-number"
    Then my wallet should appear in the wallets list
    And the wallets list should show exactly 1 result
    And I take a screenshot "wallet-visible-filtered-phone-number"

    When I navigate to the botanist wallets page
    And I filter the wallets list by my wallet address
    And I take a screenshot "filter-applied-wallet-address"
    Then my wallet should appear in the wallets list
    And the wallets list should show exactly 1 result
    And I take a screenshot "wallet-visible-filtered-wallet-address"

    When I navigate to the botanist wallets page
    And I filter the wallets list by my provider ID
    And I take a screenshot "filter-applied-provider-id"
    Then my wallet should appear in the wallets list
    And the wallets list should show exactly 1 result
    And I take a screenshot "wallet-visible-filtered-provider-id"

    When I navigate to the botanist wallets page
    And I filter the wallets list by all filters
    And I take a screenshot "filter-applied-all"
    Then my wallet should appear in the wallets list
    And the wallets list should show exactly 1 result
    And I take a screenshot "wallet-visible-filtered-all"

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

  @botanist @wallets-totp-reset @xago
  Scenario: Admin can reset a wallet user's sms otp verification
    Given that my "country" is "South Africa"
    And I completed the signup workflow
    And I completed the account verification workflow
    And I finished the phone confirmation workflow
    And I take a screenshot "phone-verified"
    And I finished the TOTP registration workflow
    And I finished the wallet address creation workflow
    When I navigate to the admin portal
    And I navigate to my wallet profile page in the admin portal
    Then the reset sms otp button should be visible
    When I click the reset sms otp button
    Then the sms otp reset confirmation modal should be visible
    And I take a screenshot "reset-sms-otp-modal"
    When I confirm the sms otp reset
    Then the reset sms otp button should not be visible
    And my SMS OTP should be not verified
    And an sms otp reset audit log entry should exist
    And I take a screenshot "reset-sms-otp-complete"
    When I start a new browser session
    And I navigate to the login page
    And I fill in the login form with my details
    And I submit the login
    Then I should be navigated to the TOTP page
    When I type in my generated totp
    And I submit the totp registration
    Then I should be navigated to the SMS OTP page
    And I take a screenshot "sms-otp-reverification-after-admin-reset"
