package workflows

type CreateLinkedCardArgs struct {
	ID         string
	WalletID   string
	ProviderID string
	Mask       string
	Name       string
	Nickname   string
}

type CreateExternalCardArgs struct {
	WalletID       string
	Name           string
	CardNumber     string
	CVV            string
	ExpirationDate string
}
