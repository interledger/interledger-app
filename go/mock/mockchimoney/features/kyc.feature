Feature: KYC Widget
  As the Protea frontend
  I want the KYC widget page to be served in an iframe
  So that users can complete identity verification without leaving the app

  # The backend constructs the KYC URL as:
  #   {CHIMONEY_KYC_BASE_URL}/verify/kyc/{externalID}?redirect={callbackURL}
  # and returns it to Protea, which renders it in an <iframe>.
  # The mock serves this page directly; it does NOT call /sub-account/kyc/link.

  Background:
    Given MockChimoney is running
    And I authenticate with a valid API key

  # ── Widget page ───────────────────────────────────────────────────────────

  Scenario: KYC page is served for a valid sub-account ID
    Given a sub-account exists with ID "kyc-sub-001"
    When I GET /verify/kyc/kyc-sub-001?redirect=https://app.test/callbacks/chimoney%3Fkyc
    Then the response status is 200
    And the Content-Type is "text/html"
    And the body contains a form for completing KYC
    And the body contains an "Approve KYC" action
    And the body contains a "Decline KYC" action

  Scenario: KYC page returns 404 for an unknown sub-account ID
    When I GET /verify/kyc/does-not-exist?redirect=https://app.test/callbacks/chimoney
    Then the response status is 404

  Scenario: KYC page requires a redirect query parameter
    Given a sub-account exists with ID "kyc-sub-002"
    When I GET /verify/kyc/kyc-sub-002 without a redirect parameter
    Then the response status is 400
    And the error message mentions "redirect"

  # ── Approval flow ─────────────────────────────────────────────────────────

  Scenario: Approving KYC redirects the browser to the redirect URL
    Given a sub-account exists with ID "kyc-sub-003"
    And the redirect URL is "https://app.test/callbacks/chimoney?kyc"
    When I POST to the KYC approval endpoint for kyc-sub-003
    Then the response redirects to a URL starting with "https://app.test/callbacks/chimoney?kyc"

  Scenario: Approving KYC updates the sub-account KYC status to completed
    Given a sub-account exists with ID "kyc-sub-004"
    When I approve KYC for kyc-sub-004
    Then the sub-account kyc status is "completed"

  Scenario: user.kyc.completed webhook fires after approval
    Given a sub-account exists with ID "kyc-sub-005"
    And a webhook receiver is listening
    When I approve KYC for kyc-sub-005
    And I wait for webhook delivery
    Then the webhook receiver received a request with body:
      | eventType | user.kyc.completed |
      | userID    | kyc-sub-005       |
    And the webhook request includes valid svix signature headers

  # ── Decline flow ──────────────────────────────────────────────────────────

  Scenario: Declining KYC redirects the browser with a failure indicator
    Given a sub-account exists with ID "kyc-sub-006"
    And the redirect URL is "https://app.test/callbacks/chimoney?kyc"
    When I POST to the KYC decline endpoint for kyc-sub-006
    Then the response redirects to a URL starting with "https://app.test/callbacks/chimoney?kyc"
    And the redirect URL includes a failure indicator query parameter

  Scenario: Declining KYC updates the sub-account KYC status to declined
    Given a sub-account exists with ID "kyc-sub-007"
    When I decline KYC for kyc-sub-007
    Then the sub-account kyc status is "declined"

  Scenario: user.kyc.declined webhook fires after rejection
    Given a sub-account exists with ID "kyc-sub-008"
    And a webhook receiver is listening
    When I decline KYC for kyc-sub-008
    And I wait for webhook delivery
    Then the webhook receiver received a request with body:
      | eventType | user.kyc.declined |
      | userID    | kyc-sub-008      |
    And the webhook request includes valid svix signature headers

  # ── Idempotency ───────────────────────────────────────────────────────────

  Scenario: KYC can only be completed once per sub-account
    Given a sub-account "kyc-sub-009" has already been approved
    When I POST to the KYC approval endpoint for kyc-sub-009 again
    Then the response status is 409
    And the error message indicates KYC is already completed
