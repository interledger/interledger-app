package gmt

import (
	"context"
)

type Client interface {
	StartUserOnboarding(ctx context.Context, walletID string) (Await, error)
	Authenticate3DS(ctx context.Context, args Authenticate3DSArgs) error
}
