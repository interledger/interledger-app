package limits

import "gitlab.com/fynbos/backend/currency"

type Limit struct {
	Daily   currency.Amount
	Monthly currency.Amount
	Overall currency.Amount
}

type FKType string

const (
	FKTypeClient FKType = "client"
)
