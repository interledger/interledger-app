package gmt

import (
	"context"

	"gitlab.com/fynbos/backend/currency"
	"gitlab.com/fynbos/backend/transactions"
)

const (
	ProviderName    = "gmt"
	TypeBankAccount = "bankAccount"
	TypeSendCard    = "sendCard"
)

type TransfersArgs struct {
	FromForeignID       string // ID against which to create transfers
	ToForeignID         string // ID against which to create transfers
	FromPaymentPointer  string // Fully qualified payment pointer
	ToPaymentPointer    string // Fully qualified payment pointer
	FromLinkedAccountID string `validate:"uuid"`
	ToLinkedAccountID   string `validate:"uuid"`
	FromWalletID        string `validate:"uuid"`
	ToWalletID          string `validate:"uuid"`
	Amount              currency.Amount
	FromTransactionID   string
}

type TransferResponse struct {
	State      transactions.State
	ExternalID string
}

type Await func(context.Context, interface{}) error
