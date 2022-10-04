package machnet

const ProviderName = "machnet"

type User struct {
	ID        string `db:"id"`
	WalletID  string `db:"wallet_id"`
	CreatedAt string `db:"created_at"`
	UpdatedAt string `db:"updated_at"`
}

type CreateArgs struct {
	WalletID   string
	ExternalID string
}

type WidgetToken struct {
	Value            string
	ExpiresInMinutes int
}
