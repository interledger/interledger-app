package analytics

import (
	"gitlab.com/fynbos/backend/currency"
	"gitlab.com/fynbos/backend/providers/machnet"
	"gitlab.com/fynbos/backend/transactions"
)

type IdentifyArgs struct {
	UserId    string
	Email     string
	FirstName string
	LastName  string
}

type WalletTransactionArgs struct {
	ID       string
	TrxType  transactions.TransactionType
	Provider transactions.Provider
	Amount   currency.Amount
	UserID   string
}

type MachnetKYCArgs struct {
	UserID   string
	WalletID string
	Status   machnet.KYCStatus
}

type MachnetCardAddedArgs struct {
	UserID   string
	WalletID string
	Scheme   string
}

type MachnetBankAddedArgs struct {
	UserID      string
	WalletID    string
	Institution string
}
