package xago

import "context"

type Await func(ctx context.Context, result interface{}) error

type SubAccount struct {
	ID             string `db:"id"`
	AccountID      string `db:"account_id"`
	DepositAddress string `db:"deposit_address"`
	DepositTag     string `db:"deposit_tag"`
	WalletID       string `db:"wallet_id"`
}

type Beneficiary struct {
	WalletID      string `db:"wallet_id"`
	BranchCode    string `db:"branch_code"`
	Reference     string `db:"reference"`
	Address       string `db:"address"`
	BankName      string `db:"bank_name"`
	AccountNumber string `db:"account_number"`
	Status        string `db:"status"`
	Currency      string `db:"currency"`
	ID            string `db:"id"`
	Scope         string `db:"scope"`
	Name          string `db:"name"`
}
