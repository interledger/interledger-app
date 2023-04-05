package tabapay

import (
	"context"
)

type Client interface {
	CreateCard(ctx context.Context, args CreateCardArgs) (Await, error)
	PullFromCard(ctx context.Context, args PullFromCardArgs) (string, error)
	PushToCard(ctx context.Context, args PushToCardArgs) (string, error)
	GetTransaction(ctx context.Context, id string) (*Transaction, error)
	Init3DS(ctx context.Context, args Init3DSArgs) (*Init3DSResponse, error)
	Lookup3DS(ctx context.Context, args Lookup3DSArgs) (*Lookup3DSResponse, error)
}
