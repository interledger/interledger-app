package limits

import "gitlab.com/fynbos/backend/currency"

type Limit struct {
	Daily   currency.Amount
	Monthly currency.Amount
	Overall currency.Amount
}

type LimitConfigured struct {
	Limit          Limit
	ForeignID      string
	ForeignDisplay string
	ForeignType    FKType
}

type FKType string

const (
	FKTypeClient          FKType = "client"
	FKTypeClientPublicKey FKType = "clientPublicKey"
)

type LimitType string

const (
	LimitTypeTransaction LimitType = "Transaction"
	LimitTypeDaily       LimitType = "Daily"
	LimitTypeMonthly     LimitType = "Monthly"
	LimitType6Monthly    LimitType = "6Monthly"
)
