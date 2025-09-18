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

	USDOpsAccount = "868196c3-f6b4-4920-bbfb-d1c7f6a98183"
	ZAROpsAccount = "b0944908-16e6-4ef4-8677-192165e33c59"
)

type Await func(ctx context.Context, result interface{}) error

type SubAccount struct {
	ID               string `db:"id"`
	AccountID        string `db:"account_id"`
	DepositAddress   string `db:"deposit_address"`
	DepositTag       int    `db:"deposit_tag"`
	WalletID         string `db:"wallet_id"`
	DepositReference string `db:"deposit_reference"`
	Details          []DepositDetails
}

type DepositDetails struct {
	ID            string            `db:"id"`
	WalletID      string            `db:"wallet_id"`
	SubAccountID  string            `db:"sub_account_id"`
	CurrencyCode  currency.Currency `db:"currency"`
	BankName      string            `db:"bank_name"`
	AccountName   string            `db:"account_name"`
	AccountNumber string            `db:"account_number"`
	BranchCode    string            `db:"branch_code"`
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
	Reference       string
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

type Withdrawal struct {
	ID     string
	Amount currency.Amount
	Status string
}

type InquiryLinkUpdate struct {
	AccountID string
	WalletID  string
}
