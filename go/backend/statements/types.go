package statements

type Statement struct {
	ID       string `validate:"omitempty,uuid"`
	Period   string
	WalletID string `validate:"uuid"`
	URI      string
}

type GenerateWalletStatementArgs struct {
	Name         string
	AccountID    string
	Period       string
	BalanceDate  string
	Balance      string
	Transactions []TransactionTableRow
}

type TransactionTableRow struct {
	Date        string
	Description string
	Amount      string
	RecieptID   string
}
