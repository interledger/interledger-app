package analytics

import (
	"gitlab.com/fynbos/backend/currency"
	"gitlab.com/fynbos/backend/transactions"
)

type IdentifyArgs struct {
	UserId    string
	Email     string
	FirstName string
	LastName  string
}

type WalletTransactionArgs struct {
	ID          string
	TrxType     transactions.TransactionType
	Provider    transactions.Provider
	Amount      currency.Amount
	ProviderFee currency.Amount
	UserID      string
}
