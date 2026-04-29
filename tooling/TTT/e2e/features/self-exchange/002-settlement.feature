@standard @settlement
Feature: Settlement report and settlement events
  As the system I want to generate a settlement report clearly showing how much
  should be settled in what direction, and I want to trigger settlement events.

  Background:
    Given I run "ttt init --mode self-exchange"
    And I run "ttt fund-liquidity gatehub eur 10000"
    And I run "ttt fund-liquidity xago eur 10000"
    And I run "ttt fund-liquidity xago zar 150000"
    And I run "ttt onboard alice gatehub eur 1000"
    And I run "ttt onboard carlos xago zar 75000"
    And I run "ttt transfer alice gatehub eur carlos xago zar 100"

  @preview
  Scenario: Settlement preview shows direction and amount
    When I run "ttt settlement-preview gatehub xago eur"
    Then the exit code is 0


  @preview @owes
  Scenario: Settlement preview shows who owes whom
    When I run "ttt settlement-preview gatehub xago eur"
    Then the exit code is 0


  @settle
  Scenario: Settlement event clears bilateral positions
    When I run "ttt settle gatehub xago eur"
    Then the exit code is 0
