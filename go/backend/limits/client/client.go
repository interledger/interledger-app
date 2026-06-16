package client

import (
	"context"

	"github.com/interledger/interledger-app/go/backend/currency"
	"github.com/interledger/interledger-app/go/backend/limits"
	"github.com/interledger/interledger-app/go/backend/limits/ops"
)

var _ limits.Client = client{}

type client struct {
	b ops.Backends
}

func New(b ops.Backends) *client {
	return &client{
		b: b,
	}
}

func (c client) ExceedsKYCLimits(ctx context.Context, walletID string, amount currency.Amount) (bool, limits.LimitType, error) {
	return ops.ExceedsKYCLimits(ctx, c.b, walletID, amount)
}
