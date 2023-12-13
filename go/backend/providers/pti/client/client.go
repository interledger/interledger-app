package client

import (
	"context"

	"gitlab.com/fynbos/backend/currency"
	"gitlab.com/fynbos/backend/providers/pti"
	"gitlab.com/fynbos/backend/providers/pti/ops"
)

var _ pti.Client = &Client{}

type Client struct {
	b ops.Backends
}

func New(b ops.Backends) *Client {
	return &Client{b}
}

func (c Client) CreateWallet(ctx context.Context, walletID string, currency currency.Currency) (pti.Await, error) {
	return ops.CreateWallet(ctx, c.b, pti.CreateWalletArgs{
		WalletID: walletID,
		Currency: currency,
		Nickname: "USD Balance",
		Title:    "USD Balance",
	})
}
