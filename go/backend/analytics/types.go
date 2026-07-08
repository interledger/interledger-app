package analytics

import (
	"github.com/interledger/interledger-app/go/backend/currency"
	"github.com/interledger/interledger-app/go/backend/transactions"
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
	ProviderFee *currency.Amount
	UserID      string
}
