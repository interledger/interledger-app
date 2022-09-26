package openpayments

type PaymentPointer struct {
	URL        string `db:"url" validate:"url"`
	WalletID   string `db:"wallet_id" validate:"uuid4"`
	Alias      string `db:"alias"`
	Asset      string `db:"asset" validate:"iso4217"`
	AssetScale int    `db:"scale" validate:"gt=0"`
}
