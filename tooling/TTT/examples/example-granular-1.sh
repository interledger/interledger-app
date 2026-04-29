# This example expands the higher-level self-exchange scenario into explicit
# account creation and direct ledger moves. It should land on the same balances
# as example-1.sh, but with every accounting leg spelled out.
./ttt reset
./ttt init --standard

# Create user accounts. User account creation also creates the provider plus
# matching system/liquidity accounts for that currency when needed.
./ttt create-account --provider gatehub --provider-name GateHub --user alice --currency EUR
./ttt create-account --provider gatehub --user bob --currency EUR
./ttt create-account --provider xago --provider-name Xago --user carlos --currency ZAR

# Add the extra internal accounts needed for the self-exchange path:
# - Xago also needs EUR system/liquidity accounts for EUR settlement.
# - Xago's ZAR FX account is the pass-through for the converted payout.
# - Mirrored EUR position accounts track the GateHub/Xago bilateral obligation.
./ttt create-account --provider xago --type system --currency EUR
./ttt create-account --provider xago --type liquidity --currency EUR
./ttt create-account --provider xago --type fx --currency ZAR
./ttt create-account --provider gatehub --type position --currency EUR --counterparty xago
./ttt create-account --provider xago --type position --currency EUR --counterparty gatehub

# Equivalent to:
#   fund-liquidity gatehub eur 10000
#   fund-liquidity xago eur 10000
#   fund-liquidity xago zar 150000
./ttt move --from gatehub/system          --to gatehub/liquidity       --currency EUR --amount 10000  --workflow "Fund Provider Liquidity" --step "fund gatehub EUR liquidity"
./ttt move --from xago/system             --to xago/liquidity          --currency EUR --amount 10000  --workflow "Fund Provider Liquidity" --step "fund xago EUR liquidity"
./ttt move --from xago/system             --to xago/liquidity          --currency ZAR --amount 150000 --workflow "Fund Provider Liquidity" --step "fund xago ZAR liquidity"

# Equivalent to:
#   onboard alice gatehub eur 1000
#   onboard carlos xago zar 75000
./ttt move --from gatehub/system          --to gatehub/alice           --currency EUR --amount 1000   --workflow "User Onboard"            --step "onboard alice"
./ttt move --from xago/system             --to xago/carlos             --currency ZAR --amount 75000  --workflow "User Onboard"            --step "onboard carlos"

# The fixed example rate is 1 EUR = 15 ZAR, so Alice's 100 EUR becomes
# 1500 ZAR for Carlos. The EUR leg creates the bilateral position; the ZAR leg
# pays Carlos from Xago's prefunded ZAR liquidity through Xago's FX account.
./ttt move --from gatehub/alice           --to gatehub/liquidity       --currency EUR --amount 100    --workflow "Cross-Provider Transfer" --step "debit sender user; credit sender liquidity"
./ttt move --from gatehub/liquidity       --to gatehub/position:xago   --currency EUR --amount 100    --workflow "Cross-Provider Transfer" --step "debit sender liquidity; credit sender position"
./ttt move --from xago/position:gatehub   --to xago/liquidity          --currency EUR --amount 100    --workflow "Cross-Provider Transfer" --step "debit recipient position; credit recipient liquidity"
./ttt move --from xago/liquidity          --to xago/fx                 --currency ZAR --amount 1500   --workflow "Cross-Provider Transfer" --step "debit recipient liquidity; credit recipient FX account"
./ttt move --from xago/fx                 --to xago/carlos             --currency ZAR --amount 1500   --workflow "Cross-Provider Transfer" --step "debit recipient FX account; credit recipient user"


# GateHub is owed 100 EUR, so its position is cleared into liquidity while
# Xago pays out of liquidity to clear the mirrored position.
./ttt move --from gatehub/position:xago   --to gatehub/liquidity       --currency EUR --amount 100    --workflow "Bilateral Settlement"    --step "debit creditor position; credit creditor liquidity"
./ttt move --from xago/liquidity          --to xago/position:gatehub   --currency EUR --amount 100    --workflow "Bilateral Settlement"    --step "debit debtor liquidity; credit debtor position"

# Write the spreadsheet report for this granular run.
./ttt export-ods
