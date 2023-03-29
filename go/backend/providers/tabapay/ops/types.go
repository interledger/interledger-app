package ops

import "gitlab.com/fynbos/backend/currency"

type PullFromCardArgs struct {
	WalletID            string
	ProviderID          string
	ReferenceID         string
	SettlementAccountID string
	Amount              currency.Amount
}
