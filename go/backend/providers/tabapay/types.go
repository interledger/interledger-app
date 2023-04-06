package tabapay

import (
	"context"

	"gitlab.com/fynbos/backend/currency"
)

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

type PullFromCardArgs struct {
	WalletID    string
	ProviderID  string
	ReferenceID string
	Amount      currency.Amount
}

type PushToCardArgs = PullFromCardArgs

type Transaction struct {
	ID             string
	ReferenceID    string
	Status         string
	OriginalStatus string
	Amount         currency.Amount
	ReversalStatus string
}

type Await func(ctx context.Context, result interface{}) error

type (
	Init3DSArgs struct {
		Amount            currency.Amount
		OutgoingPaymentID string
		CardID            string
	}

	Init3DSResponse struct {
		ID                  string
		JWT                 string
		DeviceCollectionURL string
	}
)
