Feature: HMAC signature authentication
  As a wallet integrator
  I want MockGatehub to verify request signatures using the real Gatehub format
  So that my client's signing logic is validated before hitting production

  # Signature base string format (per Gatehub docs):
  #   timestamp_ms|METHOD|full_url|body
  # - timestamp_ms: Unix timestamp in milliseconds
  # - METHOD: uppercase HTTP method
  # - full_url: absolute URL including scheme, host, path, and query string
  # - body: request body (omitted for GET or empty body — trailing pipe is trimmed)
  # Secret is used as the HMAC-SHA256 key and the result is hex-encoded.

  Background:
    Given a clean MockGatehub instance with authentication enforced
    And valid credentials with app id "local-test-app-id" and secret "local-test-app-secret"

  # ── Happy-path: full-URL signature accepted ─────────────────────────

  Scenario: POST with full-URL signature is accepted
    Given the current timestamp in milliseconds
    And the base string "timestamp_ms|POST|http://localhost:25151/auth/v1/users/managed|{body}"
    When I POST /auth/v1/users/managed with body '{"email":"sig-test@example.com"}' signed using the full URL
    Then the response status is 201
    And the response contains a user id

  Scenario: GET with full-URL signature is accepted (no body)
    Given an existing managed user
    And the current timestamp in milliseconds
    And the base string "timestamp_ms|GET|http://localhost:25151/id/v1/users/{userId}" with no body component
    When I GET /id/v1/users/{userId} signed using the full URL
    Then the response status is 200

  Scenario: POST with query parameters uses full URL including query in signature
    Given the current timestamp in milliseconds
    And the base string includes the query string in the URL component
    When I POST /auth/v1/tokens?clientId=test-client with body '{"scope":["auth"]}' signed using the full URL with query
    Then the response status is 200

  # ── Reject: path-only signature (old simple format removed) ─────────

  Scenario: POST signed with path-only URL is rejected
    Given the current timestamp in milliseconds
    And the base string "timestamp_ms|POST|/auth/v1/users/managed|{body}" using path only
    When I POST /auth/v1/users/managed with body '{"email":"path-only@example.com"}' signed using only the path
    Then the response status is 401

  Scenario: POST signed with old simple concatenation format is rejected
    Given the current timestamp in milliseconds
    And the base string "timestampPOST/auth/v1/users/managed{body}" using the old simple format
    When I POST /auth/v1/users/managed with body '{"email":"simple@example.com"}' signed using the simple format
    Then the response status is 401

  # ── Reject: missing or invalid headers ──────────────────────────────

  Scenario: Request with missing signature header is rejected
    When I POST /auth/v1/users/managed with body '{"email":"no-sig@example.com"}' without the x-gatehub-signature header
    Then the response status is 401

  Scenario: Request with missing timestamp header is rejected
    When I POST /auth/v1/users/managed with body '{"email":"no-ts@example.com"}' without the x-gatehub-timestamp header
    Then the response status is 401

  Scenario: Request with missing app-id header is rejected
    When I POST /auth/v1/users/managed with body '{"email":"no-app@example.com"}' without the x-gatehub-app-id header
    Then the response status is 401

  Scenario: Request with unknown app-id is rejected
    When I POST /auth/v1/users/managed with body '{"email":"bad-app@example.com"}' using app id "unknown-app-id"
    Then the response status is 401

  Scenario: Request signed with wrong secret is rejected
    Given the current timestamp in milliseconds
    When I POST /auth/v1/users/managed with body '{"email":"wrong-secret@example.com"}' signed with secret "wrong-secret"
    Then the response status is 401

  # ── URL reconstruction from forwarded headers ───────────────────────

  Scenario: Proxy headers are used to reconstruct the full URL for validation
    Given the current timestamp in milliseconds
    And the base string uses "https://mockgatehub.interledger.test/auth/v1/users/managed" as the URL
    When I POST /auth/v1/users/managed with body '{"email":"proxy@example.com"}' signed using the proxied URL
    And the request includes header X-Forwarded-Proto "https"
    And the request includes header X-Forwarded-Host "mockgatehub.interledger.test"
    Then the response status is 201

  # ── Public endpoints bypass authentication ──────────────────────────

  Scenario: Health endpoint does not require authentication
    When I GET /health without any HMAC headers
    Then the response status is 200

  Scenario: Root endpoint does not require authentication
    When I GET / without any HMAC headers
    Then the response status is 200

  Scenario: Iframe submit endpoint does not require authentication
    When I POST /iframe/submit without any HMAC headers
    Then the response status is not 401

  # ── Timestamp handling ──────────────────────────────────────────────
  # The x-gatehub-timestamp header value is used VERBATIM in the signature
  # base string. Gatehub and the Interledger App client both use milliseconds,
  # so MockGatehub must not normalize or convert the value before computing
  # the expected HMAC.

  Scenario: Millisecond timestamp in header is used verbatim in signature base string
    Given a fixed timestamp value "1686040166173"
    And body '{"email":"ms@example.com"}'
    And the signature is computed from "1686040166173|POST|http://localhost:25151/auth/v1/users/managed|{body}"
    When I POST /auth/v1/users/managed with that body and x-gatehub-timestamp set to "1686040166173"
    Then the response status is 201

  Scenario: Timestamp value is not normalized to seconds before signature check
    Given a fixed timestamp value "1686040166173"
    And body '{"email":"norm@example.com"}'
    And the signature is incorrectly computed using seconds "1686040166" in the base string
    When I POST /auth/v1/users/managed with that body and x-gatehub-timestamp set to "1686040166173"
    Then the response status is 401
