@standard @settlement
Feature: Settlement report and settlement events
  As the system I want to generate a settlement report clearly showing how much
  should be settled in what direction, and I want to trigger settlement events.

  Background:
    Given I run "ttt init --mode standard"
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
    And the output contains "gatehub"
    And the output contains "xago"
    And the output contains "EUR"

  @preview @owes
  Scenario: Settlement preview shows who owes whom
    When I run "ttt settlement-preview gatehub xago eur"
    Then the exit code is 0
    And the output contains "owes"

  @settle
  Scenario: Settlement event clears bilateral positions
    When I run "ttt settle gatehub xago eur"
    Then the exit code is 0
    And the output contains "Settled"

  @settle @preview
  Scenario: After settlement positions are closed
    Given I run "ttt settle gatehub xago eur"
    When I run "ttt settlement-preview gatehub xago eur"
    Then the exit code is 0
    And the output contains "Nothing to settle"

  @settle @ledger
  Scenario: Settlement recorded in ledger
    Given I run "ttt settle gatehub xago eur"
    When I run "ttt ledger"
    Then the exit code is 0
    And the output contains "Bilateral Settlement"

  @preview @no-transfers
  Scenario: Preview with no transfers returns no positions message
    Given I run "ttt reset"
    And I run "ttt init --mode standard"
    When I run "ttt settlement-preview gatehub xago eur"
    Then the exit code is 0
    And the output contains "No position accounts found"
