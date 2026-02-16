Feature: User Signup
  As a new user
  I want to sign up for an account
  So that I can access the Interledger wallet

  Background:
    Given a random test identifier is generated
    And the frontend is running at "https://interledger.test"
    Given the details of 'signup-user' are
      | field           | value                        |
      | emailSuffix     | signup@example.com           |
      | password        | InterlEdger2025!TestPassword |
      | firstName       | Signupper                    |
      | lastName        | Donaldson                    |
      | dateOfBirth     | 1995-03-20                   |
    And I impersonate 'signup-user'

  @signup @xago
  Scenario: Successfully sign up as a South-African user
    Given that my "country" is "South Africa"
    And I completed the signup workflow
    And I completed the account verification workflow
    And I finished the TOTP registration workflow
    And I finished the wallet address creation workflow
    Then I should be navigated back to the dashboard with reserved wallet status
    And I take a screenshot "signup-complete"

  @signup @gatehub
  Scenario: Successfully sign up as a German user
    Given that my "country" is "germany"
    And I completed the signup workflow
    And I completed the account verification workflow
    And I finished the TOTP registration workflow
    And I finished the wallet address creation workflow
    Then I should be navigated back to the dashboard with reserved wallet status
    And I take a screenshot "signup-complete"

  @signup @gatehub
  Scenario: Signup form validates required fields
    Given the details of 'signup-invalid-user' are
      | field           | value                        |
      | emailSuffix     | hendry@example.com           |
      | password        | InterlEdger2025!TestPassword |
      | country         | Germany                      |
      | countryCode     | DE                           |
      | firstName       | Hendry                       |
      | lastName        | Dogger                       |
      | dateOfBirth     | 1995-03-20                   |
    And I impersonate 'signup-invalid-user'
    And I navigate to the signup page
    When I click the "Sign Up" button
    Then I should see the signup form
    When I try to submit without filling required fields
    Then I should see validation errors or the form should validate on blur
    And I take a screenshot "validation-errors"