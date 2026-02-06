Feature: User KYC and Account Activation
  As a signed-up user
  I want to activate my account and complete KYC verification
  So that I can deposit funds and make payments

  Background:
    Given a random test identifier is generated
    And the frontend is running at "https://interledger.test"
    And Rafiki assets are seeded

  @kyc-min
  Scenario: Successfully activate account and complete KYC as verified user
    Given the details of 'kyc-user' are
      | field           | value                        |
      | emailSuffix     | jason@example.com            |
      | password        | InterlEdger2025!TestPassword |
      | country         | Germany                      |
      | firstName       | Jason                        |
      | lastName        | Hurry                        |
      | dateOfBirth     | 1990-01-01                   |
    And I impersonate 'kyc-user'
    And I completed the signup workflow

    # Ensure signup exists before login
    Then a signup record should exist for myself

    # Trigger verification and login
    When I trigger user verification for myself
    And I clear the browser session
    And I navigate to the login page
    And I fill in the login form with my details
    And I submit the login
    Then I should be navigated to the TOTP page
    
    # Register totp
    When I type in my generated totp
    And I submit the totp registration
    Then I should be navigated to the application dashboard
    
    # Create wallet address (required after signup)
    Then I should be redirected to the wallet address creation page
    
    When I fill in and submit the wallet address form with a unique address
    And I take a screenshot "address-create"
    And I click the "save" button on the wallet-address-form
    Then I should be navigated back to the dashboard with reserved wallet status
    And I take a screenshot "reserved"
    
    # Navigate to wallet activation
    When I navigate to the personal details page to activate wallet
    Then I should see the activate wallet button
    And I take a screenshot "details"
    
    # Trigger KYC flow and fill iframe
    When I click the "Continue" button
    And I wait for the KYC iframe to load
    And I fill and submit the mockgatehub KYC iframe
    And I wait for the KYC completion
    Then I should be navigated back to the dashboard with approved kyc status
    And I should see my account balance with kyc approved
    And I take a screenshot "kyc-completed-dashboard"    

