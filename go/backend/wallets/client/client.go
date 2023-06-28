package client

import (
	"context"

	"gitlab.com/fynbos/backend/db"
	"gitlab.com/fynbos/backend/wallets"
	"gitlab.com/fynbos/backend/wallets/ops"
)

var _ wallets.Client = &client{}

type client struct {
	b ops.Backends
}

func New(b ops.Backends) wallets.Client {
	return &client{
		b: b,
	}
}

func (c client) Create(ctx context.Context, args wallets.CreateArgs) (*wallets.Wallet, error) {
	return ops.Create(ctx, c.b, args)
}

func (c client) ForContext(ctx context.Context) (*wallets.Wallet, error) {
	return ops.WalletForContext(ctx)
}

func (c client) Get(ctx context.Context, id string) (*wallets.Wallet, error) {
	return ops.Get(ctx, c.b, id)
}

func (c client) List(ctx context.Context, userID string) ([]wallets.Wallet, error) {
	return ops.List(ctx, c.b, userID)
}

func (c client) SetWalletName(ctx context.Context, id, name string) (*wallets.Wallet, error) {
	return ops.SetWalletName(ctx, c.b, id, name)
}

func (c client) ListAll(ctx context.Context, page db.Pagination) ([]wallets.Wallet, error) {
	return ops.ListAll(ctx, c.b, page)
}

func (c client) GetFromAddress(ctx context.Context, address string) (*wallets.Wallet, error) {
	return ops.GetFromAddress(ctx, c.b, address)
}

func (c client) AddAddress(ctx context.Context, id, address string) (*wallets.Wallet, error) {
	return ops.AddAddress(ctx, c.b, id, address)
}
