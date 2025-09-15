package transactions

import (
	"database/sql"
	"time"

	"gitlab.com/fynbos/backend/currency"
)

type TransactionType string

const (
	TransactionTypeOpenPaymentIncoming     TransactionType = "open_payments_incoming"
	TransactionTypeOpenOutgoingPayment     TransactionType = "open_payments_outgoing"
	TransactionTypeReceived                TransactionType = "received"
	TransactionTypeSent                    TransactionType = "sent"
	TransactionTypeWithdrawal              TransactionType = "withdrawal"
	TransactionTypeTransfer                TransactionType = "transfer"
	TransactionTypeDeposit                 TransactionType = "deposit"
	TransactionTypeWebMonetizationIncoming TransactionType = "web_monetization_incoming"
	TransactionTypeWebMonetizationOutgoing TransactionType = "web_monetization_outgoing"
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
	ProviderOpenPayments   Provider = "open_payments"
	ProviderPaymentsEngine Provider = "payments_engine"
	ProviderXago           Provider = "xago"
)

type TransferType string

const (
	TransferTypeDebitCard            TransferType = "debit_card"
	TransferTypeCreditCard           TransferType = "credit_card"
	TransferTypeDebitBankAccount     TransferType = "debit_bank_acc"
	TransferTypeCreditBankAccount    TransferType = "credit_bank_acc"
	TransferTypeDebitWebMonetization TransferType = "debit_web_monetization"
	TransferTypeDebitBalance         TransferType = "debit_balance"
	TransferTypeCreditBalance        TransferType = "credit_balance"
)

type RefundState int16

const (
	RefundStateNone RefundState = iota
	RefundStatePending
	RefundStateCompleted
)

type CreateTransactionArgs struct {
	ID                      string          `validate:"omitempty,uuid"`
	WalletID                string          `validate:"uuid"` // Fynbos wallet ID
	ForeignID               string          `validate:"omitempty"`
	ForeignType             TransactionType `validate:"required"`
	Provider                Provider        `validate:"required"`
	State                   State           `validate:"required"`
	Note                    string
	Source                  string // Usually the sending payment pointer
	Destination             string // Usually the receiving payment pointer
	Amount                  currency.Amount
	ProviderFee             *currency.Amount
	Transfers               []TransferArgs `validate:"omitempty,dive"`
	GrantID                 string
	LinkedAccountTitle      string
	DestinationIdentity     string
	DestinationIdentityType string `validate:"omitempty,oneof=Twitter Slack Discord wallet WalletID WalletURL ExternalWalletURL"`
	Reference               string
	Title                   string
}

type UpdateTransactionArgs struct {
	WalletID        string `validate:"uuid"` // InterledgerApp wallet ID
	ForeignID       string `validate:"uuid"`
	State           State  `validate:"required"`
	Amount          currency.Amount
	ProviderFee     currency.Amount
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
	Title                   string
	Note                    string
	Type                    TransactionType
	Timestamp               time.Time
	Provider                Provider
	State                   State
	Amount                  currency.Amount
	ProviderFee             *currency.Amount
	LinkedAccountTitle      string
	DestinationIdentity     string
	DestinationIdentityType string
	Reference               string
	RefundState             RefundState
}

// Transfer is the underlying transfers that make up a single Transactions
type Transfer struct {
	ID              string
	LinkedAccountID string
	ForeignID       string
	Type            TransferType
	Amount          currency.Amount
	ProviderFee     *currency.Amount
	State           State
	Timestamp       time.Time
}

// Filters to use when listing transactions.
type TransactionRangeFilter struct {
	StartTimestamp time.Time
	EndTimestamp   time.Time
}

type DbTransaction struct {
	ID                      string          `db:"id"`
	ForeignID               sql.NullString  `db:"foreign_id"`
	WalletID                sql.NullString  `db:"wallet_id"`
	ReferenceID             sql.NullString  `db:"reference_id"`
	Type                    TransactionType `db:"type"`
	State                   State           `db:"state"`
	Provider                Provider        `db:"provider"`
	Note                    sql.NullString  `db:"note"`
	Source                  sql.NullString  `db:"source"`
	Destination             sql.NullString  `db:"destination"`
	Title                   sql.NullString  `db:"title"`
	Amount                  uint64          `db:"amount"`
	ProviderFee             uint64          `db:"provider_fee"`
	Scale                   int             `db:"asset_scale"`
	Asset                   string          `db:"asset_code"`
	Timestamp               time.Time       `db:"updated_at"`
	LinkedAccountTitle      sql.NullString  `db:"linked_account_title"`
	DestinationIdentityType sql.NullString  `db:"destination_identity_type"`
	DestinationIdentity     sql.NullString  `db:"destination_identity"`
	Reference               sql.NullString  `db:"reference"`
	RefundState             RefundState     `db:"refund_state"`
}
