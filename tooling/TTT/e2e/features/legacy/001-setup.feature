@legacy @setup
Feature: Legacy mode account seeding
  As a TTT user I want to use the CLI to select legacy mode
  so that the legacy account topology is created.

  @init
  Scenario: Initialize legacy mode
    When I run "ttt init --mode legacy"
    Then the exit code is 0
    And the output contains "GateHub"
