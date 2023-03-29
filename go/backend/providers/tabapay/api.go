package tabapay

import (
	"context"
)

type Client interface {
	CreateCard(ctx context.Context, args CreateCardArgs) (Await, error)
	PullFromCard(ctx context.Context, args PullFromCardArgs) (string, error)
	PushToCard(ctx context.Context, args PushToCardArgs) (string, error)
	GetTransaction(ctx context.Context, id string) (*Transaction, error)
}
