@legacy @cross-provider
Feature: Legacy mode cross-provider transfers
  In legacy mode only GateHub has EUR accounts and only Xago has ZAR accounts.
  Transfers use single-POS routing: the EUR obligation is tracked as a one-sided
  position on GateHub's EUR liquidity; Xago services ZAR payouts from its system account.

  Background:
    Given I run "ttt init --mode legacy"
    And I run "ttt fund-liquidity gatehub eur 10000"
    And I run "ttt onboard alice gatehub eur 100"
    And I run "ttt onboard bob gatehub eur 500"

  @gatehub-to-xago
  Scenario: GateHub EUR user sends to Xago ZAR account
    When I run "ttt transfer alice gatehub eur carlos xago zar 10"
    Then the exit code is 0
    And the output contains "alice"
    And the output contains "carlos"
    And the output contains "EUR"
    And the output contains "ZAR"

  @xago-to-gatehub
  Scenario: Xago ZAR user sends back to GateHub EUR account
    Given I run "ttt transfer alice gatehub eur carlos xago zar 10"
    When I run "ttt transfer carlos xago zar alice gatehub eur 5"
    Then the exit code is 0
    And the output contains "carlos"
    And the output contains "alice"
    And the output contains "ZAR"
    And the output contains "EUR"

  @integrity
  Scenario: Global balance is clean after forward transfer
    Given I run "ttt transfer alice gatehub eur carlos xago zar 10"
    When I run "ttt status"
    Then the exit code is 0
    And the output contains "Global balance:           OK"
    And the output contains "Liquidity decomposition:  OK"

  @integrity
  Scenario: Global balance is clean after both directions
    Given I run "ttt transfer alice gatehub eur carlos xago zar 10"
    And I run "ttt transfer carlos xago zar alice gatehub eur 5"
    When I run "ttt status"
    Then the exit code is 0
    And the output contains "Global balance:           OK"
    And the output contains "Liquidity decomposition:  OK"

  @position
  Scenario: GateHub holds a one-sided EUR position after forward transfer
    Given I run "ttt transfer alice gatehub eur carlos xago zar 10"
    When I run "ttt status"
    Then the exit code is 0
    And the output contains "position(EUR)@xago"

  @position
  Scenario: Reverse transfer reduces the hub position
    Given I run "ttt transfer alice gatehub eur carlos xago zar 10"
    And I run "ttt transfer carlos xago zar alice gatehub eur 5"
    When I run "ttt settlement-preview gatehub xago eur"
    Then the exit code is 0
    And the output contains "owes"

  @fx-rate
  Scenario: Transfer output shows the FX rate used
    When I run "ttt transfer alice gatehub eur carlos xago zar 10"
    Then the exit code is 0
    And the output contains "@"
    And the output contains "15"
