# Exit on error
set -e

# This example expands the higher-level self-exchange scenario into explicit
# account creation and direct ledger moves. It should land on the same balances
# as example-1.sh, but with every accounting leg spelled out.
./ttt reset
./ttt init --mode standard

# In this example we create a separate provider dedicated to doing self exchange
./ttt create-account --provider selfexchange --provider-name SelfExchange --type system --currency EUR
./ttt create-account --provider selfexchange --provider-name SelfExchange --type system --currency ZAR
./ttt create-account --provider selfexchange --provider-name SelfExchange --type fx --currency ZAR
./ttt create-account --provider selfexchange --provider-name SelfExchange --type fx --currency EUR

# # Add the extra internal accounts needed for the self-exchange path:
# # - Xago also needs EUR system/liquidity accounts for EUR settlement.
# # - Xago's ZAR FX account is the pass-through for the converted payout.
# # - Mirrored EUR position accounts track the GateHub/Xago bilateral obligation.
# ./ttt create-account --provider xago --type system --currency EUR
# ./ttt create-account --provider xago --type liquidity --currency EUR
# ./ttt create-account --provider xago --type fx --currency ZAR
./ttt create-account --provider gatehub --type position --currency EUR --counterparty xago
./ttt create-account --provider xago --type position --currency EUR --counterparty gatehub

# Fund liquidity
./ttt move --from selfexchange/system     --to selfexchange/fx         --currency EUR --amount 10000   --workflow "Onboard Provider Liquidity" --step "fund selfexchange EUR liquidity"
./ttt move --from selfexchange/system     --to selfexchange/fx         --currency ZAR --amount 10000   --workflow "Onboard Provider Liquidity" --step "fund selfexchange ZAR liquidity"

# Equivalent to:
#   onboard alice gatehub eur 1000
#   onboard carlos xago zar 75000
./ttt move --from gatehub/system          --to gatehub/alice           --currency EUR --amount 1000   --workflow "User Onboard"            --step "deposit by alice"
./ttt move --from xago/system             --to xago/carlos             --currency ZAR --amount 75000  --workflow "User Onboard"            --step "deposit by carlos"

# The fixed example rate is 1 EUR = 15 ZAR, so Alice's 100 EUR becomes
# 1500 ZAR for Carlos. The EUR leg creates the bilateral position; the ZAR leg
# pays Carlos from Xago's prefunded ZAR liquidity through Xago's FX account.
./ttt move --from gatehub/alice           --to gatehub/liquidity       --currency EUR --amount 100    --workflow "Cross-Provider Transfer" --step "take from alice"
./ttt move --from gatehub/liquidity       --to gatehub/position:xago   --currency EUR --amount 100    --workflow "Cross-Provider Transfer" --step "credit sender position"
./ttt move --from xago/position:gatehub   --to xago/liquidity          --currency EUR --amount 100    --workflow "Cross-Provider Transfer" --step "debit recipient position"

# Now do the exchange using the external selfexchange provider
./ttt move --from xago/liquidity          --to xago/system             --currency EUR --amount 100    --workflow "Cross-Provider Transfer" --step "euro leaving for exchange"
./ttt move --from selfexchange/system     --to selfexchange/fx         --currency EUR --amount 100    --workflow "Cross-Provider Transfer" --step "please exchange euro for zar"
./ttt move --from selfexchange/fx         --to selfexchange/system     --currency ZAR --amount 1500   --workflow "Cross-Provider Transfer" --step "here you go, 1500 zar"
./ttt move --from xago/system             --to xago/liquidity          --currency ZAR --amount 1500   --workflow "Cross-Provider Transfer" --step "zar coming back from exchange"

./ttt move --from xago/liquidity          --to xago/carlos             --currency ZAR --amount 1500   --workflow "Cross-Provider Transfer" --step "carlos receives zar"

./ttt settle gatehub xago eur

# Write the spreadsheet report for this granular run.
./ttt export-ods output/dedicated-exchange.ods
