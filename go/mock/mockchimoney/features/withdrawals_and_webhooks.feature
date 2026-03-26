Feature: Interac Withdrawal Payouts and Webhooks
  As the backend service
  I want to initiate Interac withdrawal payouts, check their status,
  and receive the resulting webhooks so that user withdrawals are processed end-to-end

  Background:
    Given MockChimoney is running
    And I authenticate with a valid API key
    And a sub-account exists with ID "chi-sub-002"
    And a webhook receiver is listening

  # ── Initiate withdrawal ───────────────────────────────────────────────────

  Scenario: Initiate an Interac withdrawal
    When I POST /v0.2.4/payouts/interac with body:
      | subAccount          | chi-sub-002                            |
      | debitCurrency       | CAD                                    |
      | turnOffNotification | true                                   |
      | interacs[0].name    | Alice Smith                            |
      | interacs[0].email   | alice@example.com                      |
      | interacs[0].amount  | 95.00                                  |
      | interacs[0].narration | Withdrawal via Interac               |
    Then the response status is 200
    And the response JSON "status" is "success"
    And the response data array contains 1 payout
    And each payout has an "issueID" matching the pattern "{subAccountID}_{uuid}"
    And each payout has a "chiref" field

  Scenario: Withdrawal requires debitCurrency
    When I POST /v0.2.4/payouts/interac without debitCurrency and body:
      | subAccount        | chi-sub-002       |
      | interacs[0].name  | Alice Smith       |
      | interacs[0].email | alice@example.com |
      | interacs[0].amount | 50.00            |
    Then the response status is 400
    And the error message mentions "debitCurrency"

  Scenario: Withdrawal requires at least one interac entry
    When I POST /v0.2.4/payouts/interac with body:
      | subAccount    | chi-sub-002 |
      | debitCurrency | CAD         |
    Then the response status is 400
    And the error message mentions "interacs"

  Scenario: Withdrawal for a non-existent sub-account is rejected
    When I POST /v0.2.4/payouts/interac with body:
      | subAccount          | does-not-exist    |
      | debitCurrency       | CAD               |
      | interacs[0].name    | Bob               |
      | interacs[0].email   | bob@example.com   |
      | interacs[0].amount  | 10.00             |
    Then the response status is 400
    And the response JSON "status" is "error"

  Scenario: Each interac entry in the array produces a separate payout record
    When I POST /v0.2.4/payouts/interac with two interac entries:
      | name  | email             | amount |
      | Alice | alice@example.com | 50.00  |
      | Bob   | bob@example.com   | 30.00  |
    Then the response data array contains 2 payouts
    And each payout has a distinct "issueID"

  # ── Check payout status ───────────────────────────────────────────────────

  Scenario: Check payout status by chiRef
    Given I have initiated an Interac withdrawal and recorded the chiRef
    When I POST /v0.2.4/payouts/status with body:
      | chiRef              | <the chiRef>  |
      | subAccount          | chi-sub-002   |
      | turnOffNotification | true          |
    Then the response status is 200
    And the response JSON "status" is "success"
    And the payout data "status" is "pending"
    And the payout data "type" is "interac"
    And the payout data "amount" equals the withdrawal amount

  Scenario: Check payout status requires chiRef
    When I POST /v0.2.4/payouts/status with an empty body
    Then the response status is 400
    And the error message mentions "chiRef"

  Scenario: Check payout status for unknown chiRef returns 404
    When I POST /v0.2.4/payouts/status with body:
      | chiRef | non-existent-chi-ref |
    Then the response status is 404
    And the response JSON "status" is "error"

  Scenario: Payout status reflects completed after webhook fires
    Given I have initiated a withdrawal and the payout.interac.completed webhook has been delivered
    When I POST /v0.2.4/payouts/status with the chiRef
    Then the payout data "status" is "completed"

  # ── Withdrawal webhooks ───────────────────────────────────────────────────

  Scenario: payout.interac.completed webhook fires after a withdrawal
    Given I have initiated an Interac withdrawal
    When I wait for webhook delivery
    Then the webhook receiver received a request with body fields:
      | eventType        | payout.interac.completed |
      | status           | completed                |
    And the webhook body "issueID" matches the withdrawal issueID
    And the webhook body "meta.issuer" equals the sub-account ID
    And the webhook body "meta.currency" is "CAD"
    And the webhook body "meta.amount" equals the withdrawal amount
    And the webhook request includes valid svix signature headers

  Scenario: Withdrawal webhook issueID encodes the sub-account ID
    Given I have initiated a withdrawal and waited for webhook delivery
    Then the webhook "issueID" starts with "chi-sub-002_"

  @wip
  Scenario: payout.interac.expired webhook can be configured instead of completed
    Given MockChimoney is configured to send "expired" withdrawal outcomes
    When I initiate an Interac withdrawal
    And I wait for webhook delivery
    Then the webhook receiver received a request with:
      | eventType | payout.interac.expired |

  @wip
  Scenario: payout.interac.cancelled webhook can be configured
    Given MockChimoney is configured to send "cancelled" withdrawal outcomes
    When I initiate an Interac withdrawal
    And I wait for webhook delivery
    Then the webhook receiver received a request with:
      | eventType | payout.interac.cancelled |
