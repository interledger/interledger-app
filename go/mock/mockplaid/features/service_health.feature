Feature: Service health

  Scenario: Health endpoint responds ok
    Given mockplaid is running
    When I GET "/health"
    Then the response status is 200
    And the response field "status" equals "ok"
