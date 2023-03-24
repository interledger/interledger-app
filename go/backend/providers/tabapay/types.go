package tabapay

var (
	ProviderName = "tabapay"
	TypeCard     = "card"
)

type CreateCardArgs struct {
	ReferenceID    string
	WalletID       string
	Name           string
	CardNumber     string
	CVV            string
	ExpirationDate string
}
