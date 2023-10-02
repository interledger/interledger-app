package workflows

type CreateExternalCardArgs struct {
	WalletID            string
	Name                string
	CardNumber          string
	CVV                 string
	ExpirationDate      string
	RejectDuplicateCard bool
	ReferenceID         string
}

type QueryCard struct {
	WalletID       string
	CardNumber     string
	ExpirationDate string
	CVV            string
	AVS            bool
}
