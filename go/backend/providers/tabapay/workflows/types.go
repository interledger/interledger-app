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
	LinkedAccountID string
	WalletID        string
	Name            string
	CardNumber      string
	CVV             string
	ExpirationDate  string
}
