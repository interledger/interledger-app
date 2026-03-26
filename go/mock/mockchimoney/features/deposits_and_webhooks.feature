Feature: Deposit Payments and Webhooks
  As the backend service and Protea frontend
  I want to initiate deposit payment links, simulate Interac funding via a browser pay page,
  and receive the resulting webhooks so that user deposits are processed end-to-end

  Background:
    Given MockChimoney is running
    And I authenticate with a valid API key
    And a sub-account exists with ID "chi-sub-001"
    And a webhook receiver is listening

  # ── Initiate ─────────────────────────────────────────────────────────────

  Scenario: Initiate a deposit payment link
    When I POST /v0.2.4/payment/initiate with body:
      | amount              | 100.00                              |
      | currency            | CAD                                 |
      | subAccount          | chi-sub-001                         |
      | payerEmail          | payer@example.com                   |
      | redirect_url        | https://app.test/callbacks/chimoney |
      | turnOffNotification | true                                |
    Then the response status is 200
    And the response JSON "status" is "success"
    And the response data contains a "paymentLink"
    And the response data "issueID" matches the pattern "{subAccountID}_{uuid}"
    And the response data "status" is "pending"
    And the response data "chiRef" is a non-empty string

  Scenario: Initiating a deposit without payerEmail is rejected
    When I POST /v0.2.4/payment/initiate with body:
      | amount     | 100.00      |
      | currency   | CAD         |
      | subAccount | chi-sub-001 |
    Then the response status is 400
    And the response JSON "status" is "error"
    And the error message mentions "payerEmail"

  Scenario: Initiating a deposit with an unsupported currency is rejected
    When I POST /v0.2.4/payment/initiate with body:
      | amount     | 100.00            |
      | currency   | GBP               |
      | subAccount | chi-sub-001       |
      | payerEmail | payer@example.com |
    Then the response status is 400
    And the response JSON "status" is "error"
    And the error message mentions "currency"

  Scenario: Initiating a deposit in USD is accepted
    When I POST /v0.2.4/payment/initiate with body:
      | amount     | 50.00             |
      | currency   | USD               |
      | subAccount | chi-sub-001       |
      | payerEmail | payer@example.com |
    Then the response status is 200
    And the response data contains a "paymentLink"

  Scenario: Initiating a deposit in NGN is accepted
    When I POST /v0.2.4/payment/initiate with body:
      | amount     | 80000.00          |
      | currency   | NGN               |
      | subAccount | chi-sub-001       |
      | payerEmail | payer@example.com |
    Then the response status is 200
    And the response data contains a "paymentLink"

  Scenario: Initiating a deposit without amount is rejected
    When I POST /v0.2.4/payment/initiate with body:
      | currency   | CAD               |
      | subAccount | chi-sub-001       |
      | payerEmail | payer@example.com |
    Then the response status is 400
    And the error message mentions "amount"

  Scenario: Initiating a deposit for a non-existent sub-account is rejected
    When I POST /v0.2.4/payment/initiate with body:
      | amount     | 100.00            |
      | currency   | CAD               |
      | subAccount | does-not-exist    |
      | payerEmail | payer@example.com |
    Then the response status is 400
    And the response JSON "status" is "error"

  # ── Verify ───────────────────────────────────────────────────────────────

  Scenario: Verify a pending payment before the pay page is visited
    Given I have initiated a deposit for chi-sub-001 and recorded the issueID
    When I POST /v0.2.4/payment/verify with body:
      | id         | <the issueID> |
      | subAccount | chi-sub-001   |
    Then the response status is 200
    And the response JSON "status" is "success"
    And the payment data "status" is "pending"

  Scenario: Verify payment with unknown issueID returns 404
    When I POST /v0.2.4/payment/verify with body:
      | id | non-existent-issue-id |
    Then the response status is 404
    And the response JSON "status" is "error"

  Scenario: Verify payment without id returns 400
    When I POST /v0.2.4/payment/verify with an empty body
    Then the response status is 400

  # ── Pay page ─────────────────────────────────────────────────────────────

  Scenario: Pay page is served as HTML for a valid issueID
    Given I have initiated a deposit and have the paymentLink
    When I GET the paymentLink URL
    Then the response status is 200
    And the Content-Type is "text/html"
    And the body contains a form for confirming payment

  Scenario: Pay page returns 404 for an unknown issueID
    When I GET /pay/non-existent-issue-id
    Then the response status is 404

  Scenario: Completing payment on the pay page redirects to redirect_url
    Given I have initiated a deposit with:
      | redirect_url | https://app.test/callbacks/chimoney |
      | payerEmail   | payer@example.com                   |
    When I POST to the pay page confirm endpoint for the issueID
    Then the response redirects to a URL starting with "https://app.test/callbacks/chimoney"
    And the redirect URL includes query parameter "issueID" matching the issueID
    And the redirect URL includes query parameter "status" equal to "success"

  Scenario: Payment is marked redeemed after pay page completion
    Given I have completed payment via the pay page for issueID "<issueID>"
    When I POST /v0.2.4/payment/verify with:
      | id         | <the issueID> |
      | subAccount | chi-sub-001   |
    Then the payment data "status" is "redeemed"

  Scenario: Verify response after completion includes fee metadata
    Given I have completed payment via the pay page
    When I verify the payment
    Then the payment data nested field "meta.processingFee.provider" is "interac"
    And the payment data nested field "meta.processingFee.amount" is a positive number

  # ── Deposit webhooks ─────────────────────────────────────────────────────

  Scenario: charge.interac.completed webhook fires after pay page completion
    Given I have initiated a deposit and completed it via the pay page
    When I wait for webhook delivery
    Then the webhook receiver received a request with body:
      | eventType | charge.interac.completed |
      | status    | completed                |
    And the webhook body "issueID" matches the deposit issueID
    And the webhook request includes valid svix signature headers

  Scenario: chimoney.redeem.completed webhook fires after pay page completion
    Given I have initiated a deposit and completed it via the pay page
    When I wait for webhook delivery
    Then the webhook receiver received a request with body:
      | eventType | chimoney.redeem.completed |
      | status    | completed                 |
    And the webhook body "issueID" matches the deposit issueID
    And the webhook request includes valid svix signature headers

  Scenario: Both deposit webhooks are delivered in sequence
    Given I have initiated a deposit and completed it via the pay page
    When I wait for all webhooks to be delivered
    Then the webhook receiver received exactly 2 requests
    And the eventTypes received are "charge.interac.completed" and "chimoney.redeem.completed"

  Scenario: Deposit webhooks include the issueID in the expected format
    Given I have initiated a deposit and completed it via the pay page
    When I wait for webhook delivery
    Then the webhook body "issueID" starts with the sub-account ID followed by "_"

  @wip
  Scenario: chimoney.redeem.failed webhook can be triggered for a failed payment
    Given I have initiated a deposit
    When the deposit is explicitly marked as failed
    And I wait for webhook delivery
    Then the webhook receiver received a request with:
      | eventType | chimoney.redeem.failed |
      | status    | failed                 |
