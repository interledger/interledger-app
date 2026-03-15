Feature: Managed user authentication and KYC
  As a wallet integrator
  I want to manage users and their verification flow
  So that downstream payments can rely on authenticated, verified identities

  Background:
    Given a clean MockGatehub instance
    And HMAC headers using app id "local-test-app-id" and secret "local-test-app-secret"

  Scenario: Create a managed user
    When I POST /auth/v1/users/managed with email "testuser@example.com"
    Then the response status is 201
    And the payload includes a generated user id
    And the managed flag is true

  Scenario: Start KYC sets the user to action_required
    Given an existing managed user
    When I POST /id/v1/users/{userId}/hubs/gw
    Then the response contains a token for the KYC flow
    And GET /id/v1/users/{userId} shows kyc_state "action_required" and risk_level "low"

  Scenario: KYC iframe is served for onboarding
    When I GET /iframe/onboarding?token={token}&user_id={userId}
    Then the response is HTML that mentions "KYC Verification" and "MockGatehub"

  Scenario: KYC submission without 2FA proceeds normally
    Given an existing managed user
    When I POST /id/v1/users/{userId}/hubs/gw
    And I submit the KYC form for user {userId} without 2FA
    Then the response status is 200
    And GET /id/v1/users/{userId} shows kyc_state "accepted" and risk_level "low"

  Scenario: KYC submission with 2FA TOTP but no org callback URL
    Given an existing managed user
    When I POST /id/v1/users/{userId}/hubs/gw
    And I submit the KYC form for user {userId} with 2FA and code "123456"
    Then the response status is 400
