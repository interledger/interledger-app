package transactions

import "time"

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

type Amount struct {
	Value      uint64 `validate:"gt=0"`
	Asset      string `validate:"iso4217"`
	AssetScale int    `validate:"gt=0"`
}

type CreateTransactionArgs struct {
	WalletID    string          `validate:"uuid"` // Fynbos wallet ID
	ForeignID   string          `validate:"uuid"`
	ForeignType TransactionType `validate:"required"`
	Provider    Provider        `validate:"required"`
	State       State           `validate:"required"`
	Note        string
	Source      string // Usually the sending payment pointer
	Destination string // Usually the receiving payment pointer
	Amount      Amount
	Transfers   []TransferArgs `validate:"omitempty,dive"`
}

type UpdateTransactionArgs struct {
	WalletID        string `validate:"uuid"` // Fynbos wallet ID
	ForeignID       string `validate:"uuid"`
	State           State  `validate:"required"`
	Amount          Amount
	UpdateTransfers []TransferArgs `validate:"omitempty,dive"`
}

type TransferArgs struct {
	WalletID             string       `validate:"omitempty,uuid"` // Fynbos wallet ID
	TransactionForeignID string       `validate:"omitempty,uuid"`
	ForeignID            string       `validate:"uuid"`
	Type                 TransferType `validate:"required"`
	Amount               Amount
	State                State `validate:"required"`
}

// Transaction is abstract information representing either an incoming or outgoing payment, wallet top up or withdrawal.
// This is used for listing transactions for the frontend
type Transaction struct {
	ForeignID   string
	Source      string
	Destination string
	Note        string
	Type        TransactionType
	Timestamp   time.Time
	Provider    Provider
	State       State
	Amount      Amount
	Transfers   []Transfer
}

// Transfer is the underlying transfers that make up a single Transactions
type Transfer struct {
	ForeignID string
	Type      TransferType
	Amount    Amount
	State     State
	Timestamp time.Time
}
