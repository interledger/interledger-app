package client

import (
	"context"

	"gitlab.com/fynbos/backend/currency"
	"gitlab.com/fynbos/backend/limits"
	"gitlab.com/fynbos/backend/limits/ops"
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

func (c client) Exceeds(ctx context.Context, walletID, clientID string, amount currency.Amount) (bool, error) {
	return ops.Exceeds(ctx, c.b, walletID, clientID, amount)
}

func (c client) UpdateClientLimits(ctx context.Context, walletID, clientURL string, limit limits.Limit) error {
	return ops.UpdateClientLimits(ctx, c.b, walletID, clientURL, limit)
}

func (c client) UpdatePublicKeyLimits(ctx context.Context, walletID, keyUuid string, limit limits.Limit) error {
	return ops.UpdatePublicKeyLimits(ctx, c.b, walletID, keyUuid, limit)
}

func (c client) List(ctx context.Context, walletID string) ([]limits.LimitConfigured, error) {
	return ops.ListLimits(ctx, c.b, walletID)
}

func (c client) GetPublicKeyLimit(ctx context.Context, walletID, publicKeyUUid string) (*limits.Limit, error) {
	return ops.GetPublicKeyLimit(ctx, c.b, walletID, publicKeyUUid)
}

func (c client) ExceedsKYCLimits(ctx context.Context, walletID string, amount currency.Amount) (bool, limits.LimitType, error) {
	return ops.ExceedsKYCLimits(ctx, c.b, walletID, amount)
}
