Feature: Organization Configuration Management
  As an application administrator
  I want to update organization-level configuration
  So that I can control API callback settings and 2FA requirements

  Background:
    Given a clean MockGatehub instance
    And HMAC headers using app id "local-test-app-id" and secret "local-test-app-secret"

  Scenario: Successfully update organization configuration with TOTP
    When I PATCH /auth/v1/users/organization/default-org with apiBaseUrl "https://api.example.com" and type2fa "totp"
    Then the response status is 200
    And the response field "id" is "default-org"
    And the response field "apiBaseUrl" is "https://api.example.com"
    And the response field "type2fa" is "totp"
    And the response has "createdAt" and "updatedAt" timestamps

  Scenario: Successfully update organization configuration with SMS
    When I PATCH /auth/v1/users/organization/default-org with apiBaseUrl "https://api.secure.com" and type2fa "sms"
    Then the response status is 200
    And the response field "type2fa" is "sms"

  Scenario: Allow non-HTTPS API base URL so that we can test in dev
    When I PATCH /auth/v1/users/organization/default-org with apiBaseUrl "http://api.example.com" and type2fa "totp"
    Then the response status is 200

  Scenario: Reject invalid 2FA type
    When I PATCH /auth/v1/users/organization/default-org with apiBaseUrl "https://api.example.com" and type2fa "invalid"
    Then the response status is 400

  Scenario: Organization not found
    When I PATCH /auth/v1/users/organization/nonexistent-org with apiBaseUrl "https://api.example.com" and type2fa "sms"
    Then the response status is 404

  Scenario: Verify configuration persists after update
    When I PATCH /auth/v1/users/organization/default-org with apiBaseUrl "https://api.persistent.com" and type2fa "totp"
    Then the response status is 200
    When I PATCH /auth/v1/users/organization/default-org with apiBaseUrl "https://api.updated.com" and type2fa "sms"
    Then the response status is 200
    And the response field "apiBaseUrl" is "https://api.updated.com"
    And the response field "type2fa" is "sms"
