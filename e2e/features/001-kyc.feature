Feature: User KYC and Account Activation
  As a signed-up user
  I want to activate my account and complete KYC verification
  So that I can deposit funds and make payments

  Background:
    Given a random test identifier is generated
    And the frontend is running at "https://interledger.test"
    And mockgatehub is running at "https://mockgatehub.interledger.test"
    And mockxago is running at "https://mockxago.interledger.test"
    And Rafiki assets are seeded
    Given the details of 'kyc-user' are
      | field           | value                        |
      | emailSuffix     | kyc@example.com              |
      | password        | InterlEdger2025!TestPassword |
      | firstName       | Jason                        |
      | lastName        | Hurry                        |
      | dateOfBirth     | 1990-01-01                   |
    And I impersonate 'kyc-user'


  @kyc @gatehub
  Scenario: Successfully activate Germany account and complete KYC as verified user
    Given that my "country" is "germany"
    And I completed the signup workflow
    And I completed the account verification workflow
    And I finished the TOTP registration workflow
    And I finished the wallet address creation workflow

    # Shows "Complete these steps to confirm your identity and activate your wallet"
    Then I should be shown the "Activate wallet" prompt form

    # Trigger KYC flow and fill iframe
    When I click the "Continue" button
    And I wait for the KYC iframe to load
    And I fill and submit the mockgatehub KYC iframe
    And I wait for the KYC completion
    Then I should be navigated back to the dashboard with approved kyc status
    And I should see my account balance with kyc approved
    And I take a screenshot "kyc-completed-dashboard"

  @kyc @pti
  Scenario: Successfully activate USA account and complete KYC
    Given that my "country" is "United States"
    And mockpti is running at "https://mockpti.interledger.test"
    And I completed the signup workflow
    And I completed the account verification workflow
    And I finished the TOTP registration workflow
    And I finished the wallet address creation workflow

    # PTI embeds the KYC form directly - no "Continue" button, iframe is injected automatically
    When I navigate to the personal details page to activate wallet
    And I wait for the KYC iframe to load
    And I fill and submit the mockpti KYC iframe
    And I wait for the KYC completion
    Then I should be navigated back to the dashboard with approved kyc status
    And I should see my account balance with kyc approved
    And I take a screenshot "kyc-pti-completed-dashboard"


  @kyc @xago
  Scenario: Successfully complete KYC as a verified user in South Africa
    Given that my "country" is "south africa"
    And I completed the signup workflow
    And I completed the account verification workflow
    And I finished the TOTP registration workflow
    And I finished the wallet address creation workflow

    # Shows "Complete these steps to confirm your identity and activate your wallet"
    Then I should be shown the "Activate wallet" prompt form

    # Trigger KYC flow and fill MockXago Persona iframe
    When I click the "Continue" button
    And I wait for the KYC iframe to load
    And I fill and submit the mockxago KYC iframe
    And I wait for the KYC completion
    Then I should be navigated back to the dashboard with approved kyc status
    And I should see my account balance with kyc approved
    And I take a screenshot "kyc-completed-dashboard"
