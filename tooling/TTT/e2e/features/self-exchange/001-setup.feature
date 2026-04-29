@self-exchange @setup
Feature: Self-exchange mode account seeding
  As a TTT user I want to use the CLI to select self-exchange mode
  so that the self-exchange account topology is created.

  @init
  Scenario: Initialize self-exchange mode
    When I run "ttt init --mode self-exchange"
    Then the exit code is 0
    And the output contains "Self-exchange"
