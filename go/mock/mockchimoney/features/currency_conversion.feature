Feature: Currency Conversion
  As the backend service
  I want to convert a local currency amount to its USD equivalent
  So that currency values can be displayed and compared in a common unit

  Background:
    Given MockChimoney is running
    And I authenticate with a valid API key

  Scenario: Convert CAD amount to USD
    When I GET /v0.2.4/info/convert/local-amount-to-usd with query params:
      | originCurrency         | CAD |
      | amountInOriginCurrency | 100 |
    Then the response status is 200
    And the response JSON "status" is "success"
    And the response data "originCurrency" is "CAD"
    And the response data "amountInOriginCurrency" is "100"
    And the response data "amountInUSD" is a positive number
    And the response data contains "validUntil"

  Scenario: Conversion result reflects the configured exchange rate
    Given MockChimoney is configured with CAD_TO_USD_RATE of "0.72"
    When I GET /v0.2.4/info/convert/local-amount-to-usd with query params:
      | originCurrency         | CAD |
      | amountInOriginCurrency | 100 |
    Then the response data "amountInUSD" is 72.0

  Scenario: Conversion requires originCurrency
    When I GET /v0.2.4/info/convert/local-amount-to-usd with query params:
      | amountInOriginCurrency | 100 |
    Then the response status is 400
    And the error message mentions "originCurrency"

  Scenario: Conversion requires amountInOriginCurrency
    When I GET /v0.2.4/info/convert/local-amount-to-usd with query params:
      | originCurrency | CAD |
    Then the response status is 400
    And the error message mentions "amountInOriginCurrency"

  Scenario: Conversion with zero amount returns zero USD
    When I GET /v0.2.4/info/convert/local-amount-to-usd with query params:
      | originCurrency         | CAD |
      | amountInOriginCurrency | 0   |
    Then the response status is 200
    And the response data "amountInUSD" is 0

  Scenario: Conversion is repeatable with the same rate
    When I convert 200 CAD to USD
    And I convert 200 CAD to USD again
    Then both responses return the same "amountInUSD"

  Scenario: Large amounts convert correctly
    When I GET /v0.2.4/info/convert/local-amount-to-usd with query params:
      | originCurrency         | CAD    |
      | amountInOriginCurrency | 100000 |
    Then the response status is 200
    And the response data "amountInUSD" is a positive number
