package transactions

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
)

type Amount struct {
	Value      uint64 `validate:"gt=0" json:"value,string"`
	Asset      string `validate:"iso4217"  json:"assetCode"`
	AssetScale int    `validate:"gt=0" json:"assetScale"`
}

type CreateTransactionArgs struct {
	WalletID    string // Fynbos wallet ID
	ForeignID   string
	ForeignType TransactionType
	Provider    Provider
	Note        string
	State       State
	Source      string // Usually the sending payment pointer
	Destination string // Usually the receiving payment pointer
	Amount      Amount
	Transfers   []CreateTransferArgs
}

type UpdateTransactionArgs struct {
	ForeignID string
	State     State
	Amount    Amount
}

type CreateTransferArgs struct {
	TransactionForeignID string
	ForeignID            string
	Type                 TransferType
	Amount               Amount
}
