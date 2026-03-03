Feature: Service health readiness
  As a developer running the test harness
  I want MockGatehub to expose a reliable health check
  So that integration flows only proceed when the service is ready

  Background:
    Given MockGatehub is started via docker-compose in the test environment
    And the service URL is http://localhost:25151

  Scenario: Health endpoint reports OK
    When I send a GET request to /health with valid HMAC headers
    Then the response status is 200
    And the response body contains "status":"ok"