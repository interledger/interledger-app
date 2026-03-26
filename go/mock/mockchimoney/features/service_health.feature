Feature: Service health and readiness
  As a developer or orchestration tool
  I want MockChimoney to expose a reliable health check endpoint
  So that dependent services and test suites only proceed when the mock is ready

  Background:
    Given MockChimoney is running

  Scenario: Health endpoint returns ok
    When I send GET /health
    Then the response status is 200
    And the response body is JSON with "status" equal to "ok"

  Scenario: Health endpoint does not require authentication
    Given authentication is enforced
    When I send GET /health without an X-API-KEY header
    Then the response status is 200
    And the response body is JSON with "status" equal to "ok"
