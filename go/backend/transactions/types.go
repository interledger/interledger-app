package transactions

import (
	"time"

	"gitlab.com/fynbos/backend/currency"
)

type TransactionType string

const (
	TransactionTypeOpenPaymentIncoming     TransactionType = "open_payments_incoming"
	TransactionTypeOpenOutgoingPayment     TransactionType = "open_payments_outgoing"
	TransactionTypeMachnetWalletTopUp      TransactionType = "machnet_wallet_topup"
	TransactionTypeMachnetWalletWithdrawal TransactionType = "machnet_wallet_withdrawal"
)

type State string

const (
	StatePending   State = "Pending"
	StateCompleted State = "Completed"
	StateFailed    State = "Failed"
)

type Provider string

const (
	ProviderOpenPayments Provider = "open_payments"
	ProviderMachnet      Provider = "machnet"
)

type TransferType string

const (
	TransferTypeDebitCard           TransferType = "debit_card"
	TransferTypeCreditMachnetWallet TransferType = "credit_wallet"
	TransferTypeDebitMachnetWallet  TransferType = "debit_wallet"
	TransferTypeCreditBankAccount   TransferType = "credit_bank_acc"
)

type CreateTransactionArgs struct {
	WalletID    string          `validate:"uuid"` // Fynbos wallet ID
	ForeignID   string          `validate:"uuid"`
	ForeignType TransactionType `validate:"required"`
	Provider    Provider        `validate:"required"`
	State       State           `validate:"required"`
	Note        string
	Source      string // Usually the sending payment pointer
	Destination string // Usually the receiving payment pointer
	Amount      currency.Amount
	Transfers   []TransferArgs `validate:"omitempty,dive"`
}

type UpdateTransactionArgs struct {
	WalletID        string `validate:"uuid"` // Fynbos wallet ID
	ForeignID       string `validate:"uuid"`
	State           State  `validate:"required"`
	Amount          currency.Amount
	UpdateTransfers []TransferArgs `validate:"omitempty,dive"`
}

type TransferArgs struct {
	WalletID             string       `validate:"omitempty,uuid"` // Fynbos wallet ID
	TransactionForeignID string       `validate:"omitempty,uuid"`
	LinkedAccountID      string       `validate:"omitempty,uuid"`
	ForeignID            string       `validate:"uuid"`
	Type                 TransferType `validate:"required"`
	Amount               currency.Amount
	State                State `validate:"required"`
}

// Transaction is abstract information representing either an incoming or outgoing payment, wallet top up or withdrawal.
// This is used for listing transactions for the frontend
type Transaction struct {
	ID          string
	ForeignID   string
	Source      string
	Destination string
	Note        string
	Type        TransactionType
	Timestamp   time.Time
	Provider    Provider
	State       State
	Amount      currency.Amount
	Transfers   []Transfer
}

// Transfer is the underlying transfers that make up a single Transactions
type Transfer struct {
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
