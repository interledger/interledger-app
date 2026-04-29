@standard @charges
Feature: Transfer charges
  As the system I want to be able to charge a fee for cross-provider transactions
  so that operational costs are covered.

  Background:
    Given I run "ttt init --mode standard"
    And I run "ttt fund-liquidity gatehub eur 10000"
    And I run "ttt fund-liquidity xago eur 10000"
    And I run "ttt fund-liquidity xago zar 150000"
    And I run "ttt onboard alice gatehub eur 1000"

  @set-charge
  Scenario: Set a transfer charge
    When I run "ttt set-charge gatehub xago 2.5"
    Then the exit code is 0
    And the output contains "2.50"

  @status
  Scenario: Status shows configured charge
    Given I run "ttt set-charge gatehub xago 2.5"
    When I run "ttt status"
    Then the exit code is 0
    And the output contains "gatehub"
    And the output contains "2.50"

  @transfer
  Scenario: Transfer applies configured charge
    Given I run "ttt set-charge gatehub xago 1"
    When I run "ttt transfer alice gatehub eur carlos xago zar 100"
    Then the exit code is 0
    And the output contains "EUR"
    And the output contains "ZAR"

  @clear-charge
  Scenario: Clear a charge
    Given I run "ttt set-charge gatehub xago 2.5"
    When I run "ttt set-charge gatehub xago"
    Then the exit code is 0
    And the output contains "cleared"
