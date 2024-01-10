package astra

import "context"

const (
	ProviderName = "astra"
	TypeCard     = "card"
)

type Await func(ctx context.Context, result interface{}) error

type CreateCardArgs struct {
	WalletID           string
	BasisTheoryTokenID string
}
