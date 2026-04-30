@legacy @settlement
Feature: Legacy mode settlement
  In legacy mode only GateHub holds a EUR position. Settlement is one-sided:
  it closes GateHub's EUR position against Xago without requiring a bilateral
  EUR mirror on the Xago side.

  Background:
    Given I run "ttt init --mode legacy"
    And I run "ttt fund-liquidity gatehub eur 10000"
    And I run "ttt onboard alice gatehub eur 100"
    And I run "ttt transfer alice gatehub eur carlos xago zar 10"

  @preview
  Scenario: Settlement preview shows the outstanding EUR obligation
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
  Scenario: Settlement event closes the one-sided position
    When I run "ttt settle gatehub xago eur"
    Then the exit code is 0
    And the output contains "Settled"

  @settle @integrity
  Scenario: Position is zero and balance is clean after settlement
    Given I run "ttt settle gatehub xago eur"
    When I run "ttt status"
    Then the exit code is 0
    And the output contains "0.00 EUR"
    And the output contains "Global balance:           OK"
    And the output contains "Liquidity decomposition:  OK"

  @settle @preview
  Scenario: Settlement preview shows nothing to settle after settlement
    Given I run "ttt settle gatehub xago eur"
    When I run "ttt settlement-preview gatehub xago eur"
    Then the exit code is 0
    And the output contains "Nothing to settle"

  @settle @ledger
  Scenario: Settlement appears in the ledger
    Given I run "ttt settle gatehub xago eur"
    When I run "ttt ledger"
    Then the exit code is 0
    And the output contains "Settlement"

  @settle @both-directions
  Scenario: Settlement after transfers in both directions
    Given I run "ttt transfer carlos xago zar alice gatehub eur 5"
    When I run "ttt settle gatehub xago eur"
    Then the exit code is 0
    And the output contains "Settled"
