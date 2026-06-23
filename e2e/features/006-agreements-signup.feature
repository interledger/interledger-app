Feature: Signup-time agreement signatures
  As a regulator-facing platform
  I want every signup to be recorded against the configured agreement set
  So that we can prove what each user accepted at sign-up time

  Background:
    Given a random test identifier is generated
    And the frontend is running at "https://interledger.test"
    And mockgatehub is running at "https://mockgatehub.interledger.test"
    And mockxago is running at "https://mockxago.interledger.test"
    Given the details of 'agreements-user' are
      | field       | value                        |
      | emailSuffix | agreements-signup@example.com |
      | password    | InterlEdger2025!TestPassword  |
      | firstName   | Agreed                        |
      | lastName    | Signer                        |
      | dateOfBirth | 1992-04-10                    |
    And I impersonate 'agreements-user'

  @agreements @signup @xago
  Scenario: A successful signup records a signature for the configured agreement
    Given that my "country" is "South Africa"
    And I completed the signup workflow
    Then a signup record should exist in the database for myself
    And an agreement signature should exist for myself for "privacy_policy-0.0.0"
    And I take a screenshot "agreement-signature-recorded"
