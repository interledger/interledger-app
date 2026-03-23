Feature: PTI user and KYC assessment lifecycle
  As backend PTI integration
  I want to create users and run assessments
  So wallet signup and KYC flows can be exercised locally

  Scenario: Create a PTI user
    Given mockpti is running
    And valid PTI headers are present
    When I POST "/users" with a valid PTI user payload
    Then the response status should be 200
    And the response should include a PTI user id

  Scenario: Get an existing PTI user
    Given an existing PTI user in mockpti
    When I GET "/users/{userId}"
    Then the response status should be 200
    And the response should include "id" equal to "{userId}"

  Scenario: Start and read KYC assessment
    Given an existing PTI user in mockpti
    And valid PTI headers are present
    When I POST "/users/assessments" with scenario id "ilf_withdrawal"
    Then the response status should be 200
    And the response should include an assessment request id
    When I GET "/users/{userId}/assessments"
    Then the response status should be 200
    And the response should include "assessment"
