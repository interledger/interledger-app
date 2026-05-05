Feature: Mock PTI service health
  As a developer
  I want a stable health endpoint
  So I can verify mockpti is running before tests

  Scenario: Health endpoint returns ok
    Given mockpti is running
    When I send a GET request to "/health"
    Then the response status should be 200
    And the response body should contain "status" as "ok"
