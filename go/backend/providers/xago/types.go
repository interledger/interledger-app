package xago

import (
	"context"

	"gitlab.com/fynbos/backend/currency"
)

var (
	ProviderName   = "xago"
	AccTypeBalance = "balance"
	AccTypeBank    = "bank_account"

	LedgerIDZAR uint32 = 9246927 // Spells xagozar on a Nokia 3310 keyboard
	LedgerIDUSD uint32 = 9246873 // Spells xagousd on a Nokia 3320 keyboard
)

type Await func(ctx context.Context, result interface{}) error

type SubAccount struct {
	ID             string `db:"id"`
	AccountID      string `db:"account_id"`
	DepositAddress string `db:"deposit_address"`
	DepositTag     int    `db:"deposit_tag"`
	WalletID       string `db:"wallet_id"`
}

type Beneficiary struct {
	WalletID  string `db:"wallet_id"`
	Reference string `db:"reference"`
	Address   string `db:"address"`
	Status    string `db:"status"`
	Currency  string `db:"currency"`
	ID        string `db:"id"`
	Scope     string `db:"scope"`
	Name      string `db:"name"`
}

type CreateBankAccountArgs struct {
	WalletID      string
	AccountNumber string
	BranchCode    string
	BankName      string
	IBAN          string
	BIC           string
}

type CreateTransactionArgs struct {
	WalletID        string
	LinkedAccountID string
	TransactionID   string
	Amount          currency.Amount
}

type CreateBalanceAccArgs struct {
	WalletID string
	Nickname string
	Title    string
	Currency currency.Currency
}

type Transaction struct {
	ID              string
	WalletID        string
	LinkedAccountID string
	TransactionID   string
	Amount          currency.Amount
}

type Balance struct {
	Total     currency.Amount
	Available currency.Amount
}
