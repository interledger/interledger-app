package transactions

import (
	"time"

	"gitlab.com/fynbos/backend/currency"
)

type TransactionType string

const (
	TransactionTypeOpenPaymentIncoming TransactionType = "open_payments_incoming"
	TransactionTypeOpenOutgoingPayment TransactionType = "open_payments_outgoing"
)

type State string

const (
	StatePending   State = "Pending"
	StateCompleted State = "Completed"
	StateFailed    State = "Failed"
	StateOnHold    State = "OnHold"
)

type Provider string

const (
	ProviderOpenPayments Provider = "open_payments"
	ProviderGMT          Provider = "gmt"
)

type TransferType string

const (
	TransferTypeDebitCard         TransferType = "debit_card"
	TransferTypeCreditCard        TransferType = "credit_card"
	TransferTypeDebitBankAccount  TransferType = "debit_bank_acc"
	TransferTypeCreditBankAccount TransferType = "credit_bank_acc"
)

type CreateTransactionArgs struct {
	ID                      string          `validate:"omitempty,uuid"`
	WalletID                string          `validate:"uuid"` // Fynbos wallet ID
	ForeignID               string          `validate:"omitempty,uuid"`
	ForeignType             TransactionType `validate:"required"`
	Provider                Provider        `validate:"required"`
	State                   State           `validate:"required"`
	Note                    string
	Source                  string // Usually the sending payment pointer
	Destination             string // Usually the receiving payment pointer
	Amount                  currency.Amount
	Transfers               []TransferArgs `validate:"omitempty,dive"`
	GrantID                 string
	LinkedAccountTitle      string
	DestinationIdentity     string
	DestinationIdentityType string `validate:"omitempty,oneof=twitter wallet"`
	Reference               string
}

type UpdateTransactionArgs struct {
	WalletID        string `validate:"uuid"` // Fynbos wallet ID
	ForeignID       string `validate:"uuid"`
	State           State  `validate:"required"`
	Amount          currency.Amount
	UpdateTransfers []TransferArgs `validate:"omitempty,dive"`
}

type UpdateForeignIDArgs struct {
	OldForeignID string `validate:"uuid"`
	NewForeignID string `validate:"uuid"`
}

type TransferArgs struct {
	LinkedAccountID string `validate:"omitempty,uuid"`
	ForeignID       string
	Type            TransferType `validate:"required"`
	Amount          currency.Amount
	State           State `validate:"required"`
}

// Transaction is abstract information representing either an incoming or outgoing payment, wallet top up or withdrawal.
// This is used for listing transactions for the frontend
type Transaction struct {
	ID                      string
	ForeignID               string
	Source                  string
	Destination             string
	Note                    string
	Type                    TransactionType
	Timestamp               time.Time
	Provider                Provider
	State                   State
	Amount                  currency.Amount
	LinkedAccountTitle      string
	DestinationIdentity     string
	DestinationIdentityType string
	Reference               string
	Transfers               []Transfer
}

// Transfer is the underlying transfers that make up a single Transactions
type Transfer struct {
	ID              string
	LinkedAccountID string
	ForeignID       string
	Type            TransferType
	Amount          currency.Amount
	State           State
	Timestamp       time.Time
}

// Filters to use when listing transactions.
type TransactionRangeFilter struct {
	StartTimestamp time.Time
	EndTimestamp   time.Time
}
