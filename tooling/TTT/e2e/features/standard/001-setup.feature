@standard @setup
Feature: Standard mode account seeding
  As a TTT user I want to use the CLI to select standard mode
  so that the standard account topology is created.

  @init
  Scenario: Initialize standard mode
    When I run "ttt init --mode standard"
    Then the exit code is 0
    And the output contains "Standard"

  @status
  Scenario: Status shows standard paradigm after init
    Given I run "ttt init --mode standard"
    When I run "ttt status"
    Then the exit code is 0
    And the output contains "Standard"
    And the output contains "gatehub"
    And the output contains "xago"
