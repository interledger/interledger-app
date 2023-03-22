package gmt

import (
	"context"
)

type Client interface {
	StartUserOnboarding(ctx context.Context, walletID string) (Await, error)
}
