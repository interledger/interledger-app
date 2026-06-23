package client

import (
	"context"

	"github.com/interledger/interledger-app/go/backend/features"
	"github.com/interledger/interledger-app/go/backend/features/ops"
)

var _ features.Client = client{}

type client struct {
	b ops.Backends
}

func New(b ops.Backends) features.Client {
	return &client{
		b: b,
	}
}

func (c client) SetFeatures(ctx context.Context, walletID string, features features.WalletFeatures) (*features.WalletFeatures, error) {
	return ops.SetFeatures(ctx, c.b, walletID, features)
}

func (c client) Features(ctx context.Context, walletID string) (*features.WalletFeatures, error) {
	return ops.Features(ctx, c.b, walletID)
}
