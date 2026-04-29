@standard @fx
Feature: FX rate tracking
  As the system I want to keep track of forex transactions performed
  and as a TTT user I want the FX rate to change between transfers.

  Background:
    Given I run "ttt init --mode standard"
    And I run "ttt fund-liquidity gatehub eur 10000"
    And I run "ttt fund-liquidity xago eur 10000"
    And I run "ttt fund-liquidity xago zar 150000"
    And I run "ttt onboard alice gatehub eur 1000"
    And I run "ttt onboard carlos xago zar 75000"

  @status
  Scenario: Status shows initial FX rate
    When I run "ttt status"
    Then the exit code is 0
    And the output contains "EUR/ZAR"
    And the output contains "15"

  @transfer
  Scenario: Transfer output includes FX rate used
    When I run "ttt transfer alice gatehub eur carlos xago zar 50"
    Then the exit code is 0
    And the output contains "15"

  @ledger
  Scenario: Ledger records cross-provider transfer workflow
    Given I run "ttt transfer alice gatehub eur carlos xago zar 50"
    When I run "ttt ledger"
    Then the exit code is 0
    And the output contains "Cross-Provider Transfer"

  @rate-change
  Scenario: FX rate can change after multiple transfers
    Given I run "ttt transfer alice gatehub eur carlos xago zar 50"
    When I run "ttt status"
    Then the exit code is 0
    And the output contains "EUR/ZAR"
