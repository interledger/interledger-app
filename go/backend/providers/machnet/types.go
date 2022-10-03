package machnet

type User struct {
	ID string
}

type CreateArgs struct {
	WalletID   string
	ExternalID string
}

type WidgetToken struct {
	Value            string
	ExpiresInMinutes int
}
