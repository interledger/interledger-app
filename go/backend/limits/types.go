package limits

import "github.com/interledger/interledger-app/go/backend/currency"

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
	LimitTypeTransaction LimitType = "LimitTransaction"
	LimitTypeDaily       LimitType = "LimitDaily"
	LimitTypeMonthly     LimitType = "LimitMonthly"
	LimitType6Monthly    LimitType = "Limit6Monthly"
	LimitTypeYearly      LimitType = "LimitYearly"
)
