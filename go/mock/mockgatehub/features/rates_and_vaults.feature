Feature: Rates and vault metadata
  As a client application
  I want to retrieve exchange rates and vault details
  So that I can display pricing and route deposits correctly

  Background:
    Given MockGatehub is running and requests include valid HMAC headers

  Scenario: Get current exchange rates
    When I GET /rates/v1/rates/current
    Then the response status is 200
    And the payload contains a counter currency
    And at least one currency rate entry besides the counter field

  Scenario: Get liquidity provider vaults
    When I GET /rates/v1/liquidity_provider/vaults
    Then the response status is 200
    And the response includes a non-empty vaults array