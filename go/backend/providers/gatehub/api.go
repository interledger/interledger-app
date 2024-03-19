package gatehub

import "context"

type Client interface {
	CreateUser(ctx context.Context, walletID string) (Await, error)
}

type Await func(ctx context.Context, result interface{}) error
