package statements

type Statement struct {
	ID       string `validate:"omitempty,uuid"`
	Period   string
	WalletID string `validate:"uuid"`
	URI      string
}
