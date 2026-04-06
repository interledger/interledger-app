Feature: Fee Estimation
  As the backend service
  I want to estimate Interac withdrawal fees before executing payouts
  So that the correct fee can be shown to the user

  # NOTE: The /info/fee-estimate endpoint is newly added to the Chimoney API
  # and was not yet deployed to production at the time of writing. The backend
  # has a fallback formula ($1.00 + 0.5% * amount) when this endpoint fails.
  # The mock must implement the endpoint regardless.

  Background:
    Given MockChimoney is running
    And I authenticate with a valid API key

  Scenario: Estimate fee for an Interac withdrawal
    When I POST /v0.2.4/info/fee-estimate with body:
      | amount    | 100.00  |
      | currency  | CAD     |
      | rail      | interac |
      | direction | payout  |
    Then the response status is 200
    And the response JSON "status" is "success"
    And the response data "totalFee" is a positive number
    And the response data "netAmount" equals amount minus totalFee
    And the response data "currency" is "CAD"
    And the response data "rail" is "interac"
    And the response data "direction" is "payout"

  Scenario: Fee estimation requires amount
    When I POST /v0.2.4/info/fee-estimate with body:
      | currency | CAD     |
      | rail     | interac |
    Then the response status is 400
    And the error message mentions "amount"

  Scenario: Fee estimation with no rail defaults to chimoney payout (currency must be USD)
    When I POST /v0.2.4/info/fee-estimate with body:
      | amount   | 50.00 |
      | currency | USD   |
    Then the response status is 200
    And the response data "totalFee" is a positive number

  Scenario: Fee estimation without rail but non-USD currency is rejected
    When I POST /v0.2.4/info/fee-estimate with body:
      | amount   | 50.00 |
      | currency | CAD   |
    Then the response status is 400
    And the error message indicates currency must be USD when rail is not specified

  Scenario: Fee direction defaults to payout when omitted
    When I POST /v0.2.4/info/fee-estimate with body:
      | amount   | 100.00  |
      | currency | CAD     |
      | rail     | interac |
    Then the response status is 200
    And the response data "direction" is "payout"

  Scenario: Fee estimation for funding direction is accepted
    When I POST /v0.2.4/info/fee-estimate with body:
      | amount    | 100.00  |
      | currency  | CAD     |
      | rail      | interac |
      | direction | funding |
    Then the response status is 200
    And the response data "direction" is "funding"

  Scenario: Fee is consistent across multiple identical requests
    When I POST /v0.2.4/info/fee-estimate twice with the same body
    Then both responses have identical "totalFee" values

  Scenario: Fee reflects configured flat fee
    Given MockChimoney is configured with INTERAC_FEE_FLAT of "2.50"
    When I POST /v0.2.4/info/fee-estimate with body:
      | amount    | 100.00  |
      | currency  | CAD     |
      | rail      | interac |
      | direction | payout  |
    Then the response data "totalFee" is 2.50

  Scenario: netAmount equals amount minus totalFee
    When I POST /v0.2.4/info/fee-estimate with body:
      | amount    | 100.00  |
      | currency  | CAD     |
      | rail      | interac |
      | direction | payout  |
    Then the response data "netAmount" is "amount" minus "totalFee"
