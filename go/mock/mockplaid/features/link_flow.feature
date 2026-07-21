Feature: Plaid mock link flow

  Scenario: Create a link token
    When I create a link token for user "u_1"
    Then the response status is 200
    And the response field "link_token" is present

  Scenario: Bank A account id is stable across selections
    Given a link token for user "u_1"
    When I select institution "ins_mock_a" account "checking"
    And I select institution "ins_mock_a" account "checking" again
    Then both selected account ids are equal

  Scenario: Bank B mints a new account id each selection
    Given a link token for user "u_1"
    When I select institution "ins_mock_b" account "checking"
    And I select institution "ins_mock_b" account "checking" again
    Then the selected account ids differ

  Scenario: Exchange public token and resolve institution
    Given a link token for user "u_1"
    When I select institution "ins_mock_a" account "checking"
    And I exchange the public token
    Then the response field "access_token" is present
    When I resolve the item institution
    Then the institution name is "Tartan Bank"
