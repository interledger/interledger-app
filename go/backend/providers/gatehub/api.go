package gatehub

import "context"

type Client interface {
	CreateUser(ctx context.Context, walletID string) (Await, error)
	GetOnboardingWidget(ctx context.Context, walletID string) (string, error)
}

type Await func(ctx context.Context, result interface{}) error
