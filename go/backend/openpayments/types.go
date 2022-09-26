package openpayments

type PaymentPointer struct {
	URL        string `db:"url" validate:"url" json:"id"`
	WalletID   string `db:"wallet_id" validate:"uuid4" json:"-"`
	Alias      string `db:"alias" json:"publicName"`
	Asset      string `db:"asset" validate:"iso4217"  json:"assetCode"`
	AssetScale int    `db:"scale" validate:"gt=0" json:"assetScale"`
}
