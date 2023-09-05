package dynamicforms

type (
	Form struct {
		ID       string      `db:"id"`
		FormID   string      `db:"form_id"`
		Data     interface{} `db:"data"`
		WalletID string      `db:"wallet_id"`
	}
	CreateFormArgs struct {
		FormID   string
		FormData interface{}
		WalletID string
	}
)
