package limits

import "gitlab.com/fynbos/backend/currency"

type WalletLimit struct {
	GrantID  string
	ClientID string
	Limits   Limit
}

type Limit struct {
	Daily   currency.Amount
	Monthly currency.Amount
	Total   currency.Amount
}

type LimitFKType string

const (
	LimitFKTypeClient LimitFKType = "client"
)
