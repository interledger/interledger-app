@generic @mode-validation
Feature: Mode validation
  As a TTT user I want invalid mode names to be rejected
  so that I know when the CLI cannot select a topology.

  @unknown-mode
  Scenario: Unknown mode is rejected
    When I run "ttt init --mode unknown-mode"
    Then the exit code is 1
    And the error output contains "unknown mode"
