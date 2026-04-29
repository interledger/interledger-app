@standard @liquidity
Feature: Liquidity funding
  As the system I want to be able to fund liquidity into provider accounts
  so that transfers and deposits can be processed.

  Background:
    Given I run "ttt init --mode standard"

  @gatehub @eur
  Scenario: Fund EUR liquidity into GateHub
    When I run "ttt fund-liquidity gatehub eur 10000"
    Then the exit code is 0
    And the output contains "gatehub"
    And the output contains "EUR"

  @xago @eur
  Scenario: Fund EUR liquidity into Xago
    When I run "ttt fund-liquidity xago eur 5000"
    Then the exit code is 0
    And the output contains "xago"
    And the output contains "EUR"

  @xago @zar
  Scenario: Fund ZAR liquidity into Xago
    When I run "ttt fund-liquidity xago zar 150000"
    Then the exit code is 0
    And the output contains "xago"
    And the output contains "ZAR"

  @status
  Scenario: Status reflects funded liquidity balances
    Given I run "ttt fund-liquidity gatehub eur 10000"
    And I run "ttt fund-liquidity xago zar 150000"
    When I run "ttt status"
    Then the exit code is 0
    And the output contains "10000"
    And the output contains "150000"
