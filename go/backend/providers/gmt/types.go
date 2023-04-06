package gmt

import (
	"context"
)

const (
	ProviderName    = "gmt"
	TypeBankAccount = "bankAccount"
	TypeSendCard    = "sendCard"
)

type Await func(context.Context, interface{}) error

type Authenticate3DSArgs struct {
	OutgoingPaymentID      string
	ThreeDSID              string
	ThreeDSVersion         string
	ProcessorTransactionID string
	DsTransactionID        string
	Status                 string
	UCAF                   string
	XID                    string
	JWT                    string
}
