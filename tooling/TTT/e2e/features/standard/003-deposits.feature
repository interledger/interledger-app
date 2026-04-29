@standard @deposits
Feature: User deposits
  As a user I want to be able to deposit funds into my account
  so that I have a balance to send and receive payments.

  Background:
    Given I run "ttt init --mode standard"
    And I run "ttt fund-liquidity gatehub eur 10000"
    And I run "ttt fund-liquidity xago zar 150000"

  @gatehub @eur @onboard
  Scenario: GateHub user deposits EUR
    When I run "ttt onboard alice gatehub eur 500"
    Then the exit code is 0
    And the output contains "alice"
    And the output contains "EUR"

  @xago @zar @onboard
  Scenario: Xago user deposits ZAR
    When I run "ttt onboard carlos xago zar 7500"
    Then the exit code is 0
    And the output contains "carlos"
    And the output contains "ZAR"

  @status
  Scenario: Status shows user balance after deposit
    Given I run "ttt onboard alice gatehub eur 500"
    When I run "ttt status"
    Then the exit code is 0
    And the output contains "500"

  @offboard
  Scenario: User can withdraw deposited funds
    Given I run "ttt onboard alice gatehub eur 500"
    When I run "ttt offboard gatehub eur alice 200"
    Then the exit code is 0
    And the output contains "alice"

  @p2p
  Scenario: Arbitrary CLI transfer operations
    Given I run "ttt onboard alice gatehub eur 1000"
    And I run "ttt onboard bob gatehub eur 500"
    When I run "ttt p2p gatehub eur alice bob 100"
    Then the exit code is 0
    And the output contains "alice"
    And the output contains "bob"
