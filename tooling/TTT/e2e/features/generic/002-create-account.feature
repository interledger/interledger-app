@generic @create-account
Feature: Create account
  As a TTT user I want to create accounts from the CLI
  so that I can model providers and users without using a seeded paradigm.

  @existing-provider
  Scenario: Create a user account on an existing provider
    Given I run "ttt init --mode standard"
    When I run "ttt create-account --provider gatehub --user superduper --currency EUR"
    Then the exit code is 0
    And the output contains "gatehub/superduper(EUR)"

  @new-provider
  Scenario: Create a provider when creating a user account
    When I run "ttt create-account --provider blue --provider-name Blue --user bob --currency EUR"
    Then the exit code is 0
    And the output contains "blue/bob(EUR)"
    And the output contains "Created provider blue"
    When I run "ttt status"
    Then the exit code is 0
    And the output contains "Blue"
    And the output contains "blue/system(EUR)"
    And the output contains "blue/liquidity(EUR)"
