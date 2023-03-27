package tabapay

import "context"

var (
	ProviderName = "tabapay"
	TypeCard     = "card"
)

type CreateCardArgs struct {
	IdempotencyKey string
	WalletID       string
	Name           string
	CardNumber     string
	CVV            string
	ExpirationDate string
}

type Await func(ctx context.Context, result interface{}) error
