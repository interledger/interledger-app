package gatehub

import "gitlab.com/fynbos/backend/currency"

const (
	ProviderName   = "gatehub"
	AccTypeBalance = "balance"

	LedgerIDEUR uint32 = 4482387 // Spells ghubeur on a Nokia 3320 keyboard
)

type Balance struct {
	Total     currency.Amount
	Available currency.Amount
}
