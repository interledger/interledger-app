@standard @cross-provider
Feature: Cross-provider transfers
  As a GateHub user I want to be able to send EUR to a Xago ZAR account
  and as a Xago user I want to be able to send ZAR to a GateHub EUR account.

  Background:
    Given I run "ttt init --mode standard"
    And I run "ttt fund-liquidity gatehub eur 10000"
    And I run "ttt fund-liquidity xago eur 10000"
    And I run "ttt fund-liquidity xago zar 150000"
    And I run "ttt onboard alice gatehub eur 1000"
    And I run "ttt onboard carlos xago zar 75000"

  @gatehub-to-xago
  Scenario: GateHub EUR user sends to Xago ZAR account
    When I run "ttt transfer alice gatehub eur carlos xago zar 100"
    Then the exit code is 0
    And the output contains "alice"
    And the output contains "carlos"
    And the output contains "EUR"
    And the output contains "ZAR"

  @xago-to-gatehub
  Scenario: Xago ZAR user sends to GateHub EUR account
    Given I run "ttt onboard bob gatehub eur 500"
    When I run "ttt transfer carlos xago zar bob gatehub eur 1500"
    Then the exit code is 0
    And the output contains "carlos"
    And the output contains "bob"
    And the output contains "ZAR"
    And the output contains "EUR"

  @fx-rate
  Scenario: Transfer output shows FX rate
    When I run "ttt transfer alice gatehub eur carlos xago zar 100"
    Then the exit code is 0
    And the output contains "@"

  @integrity
  Scenario: Cross-provider transfer passes integrity checks
    Given I run "ttt transfer alice gatehub eur carlos xago zar 100"
    When I run "ttt status"
    Then the exit code is 0
    And the output contains "Global balance:           OK"
    And the output contains "Bilateral positions:      OK"
