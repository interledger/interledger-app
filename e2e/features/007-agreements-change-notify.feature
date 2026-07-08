Feature: Agreement change notification workflow
  As an operator who publishes a new version of a user agreement
  I want every user who signed an older version to be marked notified
  So that we have an auditable record that they were informed of the change

  Background:
    Given a random test identifier is generated
    And the frontend is running at "https://interledger.test"
    And mockgatehub is running at "https://mockgatehub.interledger.test"
    And mockxago is running at "https://mockxago.interledger.test"
    Given the details of 'agreements-change-user' are
      | field       | value                              |
      | emailSuffix | agreements-change@example.com      |
      | password    | InterlEdger2025!TestPassword       |
      | firstName   | Affected                           |
      | lastName    | Signer                             |
      | dateOfBirth | 1988-09-22                         |
    And I impersonate 'agreements-change-user'

  @agreements @notify @xago
  Scenario: A new privacy_policy version notifies the existing signer
    Given that my "country" is "South Africa"
    And I completed the signup workflow
    Then an agreement signature should exist for myself for "privacy_policy-0.0.0"
    Given a new "privacy_policy" agreement version "9.9.9" is published
    When the agreement change notification workflow runs
    Then I should be marked notified for the new agreement
    And I take a screenshot "agreement-change-notified"
