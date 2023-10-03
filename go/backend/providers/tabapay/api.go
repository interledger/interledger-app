package tabapay

import (
	"context"

	"gitlab.com/fynbos/backend/currency"
)

type Client interface {
	CreateCard(ctx context.Context, args CreateCardArgs) (Await, error)
	PullFromCard(ctx context.Context, args PullFromCardArgs) (*Transaction, error)
	PushToCard(ctx context.Context, args PushToCardArgs) (*Transaction, error)
	GetTransaction(ctx context.Context, id string) (*Transaction, error)
	Init3DS(ctx context.Context, args Init3DSArgs) (*Init3DSResponse, error)
	Lookup3DS(ctx context.Context, args Lookup3DSArgs) (*Lookup3DSResponse, error)
	Authenticate3DS(ctx context.Context, args Authenticate3DSArgs) (*Authenticate3DSResponse, error)
	Get3DSSession(ctx context.Context, id string) (*ThreeDSSession, error)
	ReverseTransaction(ctx context.Context, id string, txSettled bool) error
	GetFXRate(ctx context.Context, cc currency.Currency) (*FXRate, error)
}
