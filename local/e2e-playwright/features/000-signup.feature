Feature: User Signup
  As a new user
  I want to sign up for an account
  So that I can access the Interledger wallet

  Background:
    Given a random test identifier is generated
    And the frontend is running at "https://interledger.test"

  @signuponly
  Scenario: Successfully sign up as a German user
    Given the details of 'signup-user' are
      | field           | value                        |
      | emailSuffix     | sarah@example.com            |
      | password        | InterlEdger2025!TestPassword |
      | country         | Germany                      |
      | countryCode     | DE                           |
      | firstName       | Sarah                        |
      | lastName        | Donaldson                    |
      | dateOfBirth     | 1995-03-20                   |
    And I impersonate 'signup-user'

    When I navigate to the signup page
    And I click the "Sign Up" button
    Then I should see the signup form

    # Step 1: Profile Details
    When I fill in "first name" with my "firstName"
    And I fill in "last name" with my "lastName"
    And I fill in "email" with "emailSuffix" prefixed with the random identifier
    And I select "Germany" from the country dropdown
    And I take a screenshot "profile-filled"
    And I click the "Continue" button
    Then I should be on step 2

    # Step 2: Phone Number and possibly Password (varies by flow)
    When I fill in "phone" with a unique valid German number
    And I try to fill in "password" with my "password"
    And I try to fill in "password confirmation" with my "password"
    And I take a screenshot "phone-filled"
    And I click the "Continue" button
    
    # Step 3 may have password fields if they weren't on step 2
    When I try to fill in "password" with my "password"
    And I try to fill in "password confirmation" with my "password"
    And I check the terms and conditions checkbox
    And I take a screenshot "before-submit-signup"
    And I click the "Confirm" button
    
    # Final submission
    Then the signup should be submitted

    # Verification
    Then a signup record should exist in the database for myself
    And the signup should have first name matching my "firstName"
    And the signup should have last name matching my "lastName"
    And the signup should have country code matching my "countryCode"
    And I should be able to verify the full user status

    # Step 4 trigger user verification
    When I trigger user verification for myself
    And I clear the browser session
    And I navigate to the login page
    And I fill in my login credentials
    And I take a screenshot "logging-in"
    And I submit the login 
    Then I should be navigated to the TOTP page
    
    # Step 5 register totp
    When I type in my generated totp for myself
    And I take a screenshot "totp"
    And I submit the totp registration
    Then I should be navigated to the application dashboard
    And I take a screenshot "post-signup"


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