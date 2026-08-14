Feature: Migration email job
  As an admin announcing a migration
  I want to send the announcement to a single user before sending it to everyone
  So that I can check the email is right without emailing the whole user base

  Background:
    Given a random test identifier is generated
    And the frontend is running at "https://interledger.test"
    And mockgatehub is running at "https://mockgatehub.interledger.test"
    And mockxago is running at "https://mockxago.interledger.test"
    Given the details of 'migration-email-user' are
      | field       | value                          |
      | emailSuffix | migration-email@example.com    |
      | password    | InterlEdger2025!TestPassword   |
      | firstName   | Migrating                      |
      | lastName    | Recipient                      |
      | dateOfBirth | 1990-04-11                     |
    And I impersonate 'migration-email-user'

  @migration-email @xago
  Scenario: The job sends to one user and refuses an unknown address
    Given that my "country" is "South Africa"
    And I completed the signup workflow
    When the migration email job runs for my email
    Then the migration email job should report no failures
    When the migration email job runs for an unknown address
    Then the migration email job should fail with "no user found for:"
