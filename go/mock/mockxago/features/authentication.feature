Feature: Xago Authentication
  As a wallet backend
  I want to authenticate with Xago mock service
  So that I can access protected Xago endpoints

  Background:
    Given the Xago mock service is running
    And the environment variables are set:
      | XAGO_API_PUBLIC_KEY  | test_public_key_12345  |
      | XAGO_API_SECRET      | test_secret_key_98765  |

  Scenario: Successfully obtain access token with valid credentials
    When I request a login token with valid credentials
      | policyId    | 5e2585a474b0e90012ce8ff1         |
      | publicKey   | test_public_key_12345             |
      | secretKey   | test_secret_key_98765             |
    Then I receive a successful response with status code 200
    And the response contains a valid access token
    And the token expires in 55 minutes

  Scenario: Reject login with invalid public key
    When I request a login token with invalid credentials
      | policyId    | 5e2585a474b0e90012ce8ff1         |
      | publicKey   | invalid_public_key                |
      | secretKey   | test_secret_key_98765             |
    Then I receive an error response with status code 401
    And the error message is "unauthorized"

  Scenario: Reject login with invalid secret key
    When I request a login token with invalid credentials
      | policyId    | 5e2585a474b0e90012ce8ff1         |
      | publicKey   | test_public_key_12345             |
      | secretKey   | invalid_secret_key                |
    Then I receive an error response with status code 401
    And the error message is "unauthorized"

  Scenario: Reject login with missing policy ID
    When I request a login token with missing fields
      | publicKey   | test_public_key_12345             |
      | secretKey   | test_secret_key_98765             |
    Then I receive an error response with status code 400
    And the error message contains "policyId is required"

  Scenario: Reject login with missing public key
    When I request a login token with missing fields
      | policyId    | 5e2585a474b0e90012ce8ff1         |
      | secretKey   | test_secret_key_98765             |
    Then I receive an error response with status code 400
    And the error message contains "apiPublicKey is required"

  Scenario: Reject login with missing secret key
    When I request a login token with missing fields
      | policyId    | 5e2585a474b0e90012ce8ff1         |
      | publicKey   | test_public_key_12345             |
    Then I receive an error response with status code 400
    And the error message contains "apiSecretKey is required"

  Scenario: Reuse access token across multiple requests
    Given I have obtained a valid access token
    When I use the token to create a sub-account
      | firstName   | John             |
      | lastName    | Doe              |
      | email       | john@example.com |
    Then the request succeeds with status code 200
    When I use the same token to list currencies
    Then the request succeeds with status code 200

  Scenario: Automatically refresh expired token
    Given I have obtained an access token that is about to expire
    And I attempt to use the expired token
    When I request a new login token with valid credentials
    Then I receive a successful response with status code 200
    And the new token is different from the expired token
    And the new token is valid

  Scenario: Reject requests without token
    Given I have not obtained an access token
    When I attempt to create a sub-account without a token
      | firstName   | John             |
      | lastName    | Doe              |
      | email       | john@example.com |
    Then I receive an error response with status code 401
    And the error message is "missing authorization header"

  Scenario: Reject requests with invalid token
    Given I have an invalid access token "invalid_token_xyz"
    When I attempt to create a sub-account with the invalid token
      | firstName   | John             |
      | lastName    | Doe              |
      | email       | john@example.com |
    Then I receive an error response with status code 401
    And the error message is "invalid token"
