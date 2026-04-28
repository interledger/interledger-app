Feature: PTI JWT token generation
  As Protea PTI SDK integration
  I want PTI token generation to work via backend proxy
  So SDK forms can request signed access tokens

  Scenario: Create JWT for PTI SDK form request
    Given mockpti is running
    And valid PTI headers are present
    When I POST "/auth/jwt" with url and method payload
    Then the response status should be 200
    And the response should include "accessToken"
    And the response should include "tokenType"
    And the response should include "expiresAt"
