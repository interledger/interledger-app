package dynamicforms

type (
	Form struct {
		ID       string      `db:"id"`
		FormID   string      `db:"form_id"`
		Data     interface{} `db:"data"`
		WalletID string      `db:"wallet_id"`
	}
	FormCount struct {
		FormID string `db:"form_id"`
		Count  int32  `db:"count"`
	}
	CreateFormArgs struct {
		FormID   string
		FormData interface{}
		WalletID string
	}
)
