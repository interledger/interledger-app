@generic @move
Feature: Move funds
  As a TTT user I want to move funds directly between user accounts
  so that I can model trusted same-currency account adjustments.

  @cross-provider
  Scenario: Move funds across providers in the same currency
    Given I run "ttt init --mode standard"
    And I run "ttt onboard alice gatehub eur 100"
    And I run "ttt create-account --provider xago --user dora --currency EUR"
    When I run "ttt move --from gatehub/alice --to xago/dora --currency EUR --amount 10 --workflow doing-something --step 'part 0 out of 3'"
    Then the exit code is 0
    And the output contains "Moved 10.00 EUR gatehub/alice"
    When I run "ttt ledger"
    Then the exit code is 0
    And the output contains "doing-something"
    And the output contains "part 0 out of 3"
    When I run "ttt status"
    Then the exit code is 0
    And the output contains "gatehub/alice(EUR)"
    And the output contains "90.00"
    And the output contains "xago/dora(EUR)"
    And the output contains "10.00"

  @insufficient-balance
  Scenario: Move rejects insufficient sender funds
    Given I run "ttt init --mode standard"
    And I run "ttt onboard alice gatehub eur 5"
    And I run "ttt create-account --provider xago --user dora --currency EUR"
    When I run "ttt move --from gatehub/alice --to xago/dora --currency EUR --amount 10"
    Then the exit code is 1
    And the error output contains "insufficient balance"
