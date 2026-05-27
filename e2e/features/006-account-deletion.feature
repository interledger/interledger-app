Feature: Account Deletion
  As a wallet user
  I want to request deletion of my account
  So that my data can be removed by support

  Background:
    Given a random test identifier is generated
    And the frontend is running at "https://interledger.test"
    And mockgatehub is running at "https://mockgatehub.interledger.test"
    And Rafiki assets are seeded

  @account-deletion @gatehub @ff-off
  Scenario: Feature flag off — settings link is hidden
    Given the details of 'delete-link-off-user' are
      | field       | value                        |
      | emailSuffix | delete-link-off@example.com  |
      | password    | InterlEdger2025!TestPassword |
      | country     | Germany                      |
      | firstName   | Delete                       |
      | lastName    | LinkOff                      |
      | dateOfBirth | 1988-06-15                   |
    And I complete the minimal KYC flow `delete-link-off-user`
    When I navigate to "/settings"
    Then the "Delete account" settings link should not be visible

  @account-deletion @gatehub @ff-off
  Scenario: Feature flag off — direct navigation redirects to settings
    Given the details of 'delete-route-off-user' are
      | field       | value                        |
      | emailSuffix | delete-route-off@example.com |
      | password    | InterlEdger2025!TestPassword |
      | country     | Germany                      |
      | firstName   | Delete                       |
      | lastName    | RouteOff                     |
      | dateOfBirth | 1988-06-15                   |
    And I complete the minimal KYC flow `delete-route-off-user`
    When I navigate to "/settings/delete-account"
    Then I should be redirected to "/settings"

  @account-deletion @gatehub
  Scenario: Feature flag on — settings shows the Delete account link
    Given the details of 'delete-link-on-user' are
      | field       | value                        |
      | emailSuffix | delete-link-on@example.com   |
      | password    | InterlEdger2025!TestPassword |
      | country     | Germany                      |
      | firstName   | Delete                       |
      | lastName    | LinkOn                       |
      | dateOfBirth | 1988-06-15                   |
    And I complete the minimal KYC flow `delete-link-on-user`
    And the delete-account feature is enabled for my wallet
    When I navigate to "/settings"
    Then the "Delete account" settings link should be visible

  @account-deletion @gatehub
  Scenario: Request account deletion via TOTP step-up
    Given the details of 'delete-happy-user' are
      | field       | value                        |
      | emailSuffix | delete-happy@example.com     |
      | password    | InterlEdger2025!TestPassword |
      | country     | Germany                      |
      | firstName   | Delete                       |
      | lastName    | Happy                        |
      | dateOfBirth | 1988-06-15                   |
    And I complete the minimal KYC flow `delete-happy-user`
    And the delete-account feature is enabled for my wallet
    When I navigate to "/settings/delete-account"
    And I click the destructive "Delete account" button
    And I complete the TOTP step-up challenge
    Then I should be redirected to "/settings"
    And an account-deletion request should exist for me with status "pending"

  @account-deletion @gatehub
  Scenario: Pending request hides the link and shows the indicator on settings
    Given the details of 'delete-pending-user' are
      | field       | value                        |
      | emailSuffix | delete-pending@example.com   |
      | password    | InterlEdger2025!TestPassword |
      | country     | Germany                      |
      | firstName   | Delete                       |
      | lastName    | Pending                      |
      | dateOfBirth | 1988-06-15                   |
    And I complete the minimal KYC flow `delete-pending-user`
    And the delete-account feature is enabled for my wallet
    And an account-deletion request exists for me with status "pending"
    When I navigate to "/settings"
    Then the pending account-deletion indicator should be visible
    And the "Delete account" settings link should not be visible

  @account-deletion @gatehub
  Scenario: Duplicate request — settings.delete-account redirects when one exists
    Given the details of 'delete-dup-user' are
      | field       | value                        |
      | emailSuffix | delete-dup@example.com       |
      | password    | InterlEdger2025!TestPassword |
      | country     | Germany                      |
      | firstName   | Delete                       |
      | lastName    | Dup                          |
      | dateOfBirth | 1988-06-15                   |
    And I complete the minimal KYC flow `delete-dup-user`
    And the delete-account feature is enabled for my wallet
    And an account-deletion request exists for me with status "pending"
    When I navigate to "/settings/delete-account"
    Then I should be redirected to "/settings"
