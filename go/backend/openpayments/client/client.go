package client

import (
	"context"

	"gitlab.com/fynbos/backend/openpayments"
	"gitlab.com/fynbos/backend/openpayments/ops"
)

var _ openpayments.Client = &client{}

type client struct {
	b ops.Backends
}

func New(b ops.Backends) openpayments.Client {
	return &client{
		b: b,
	}
}

func (c client) GetWalletPaymentPointer(ctx context.Context, walletID string) (*openpayments.PaymentPointer, error) {
	return ops.GetWalletPaymentPointer(ctx, c.b, walletID)
}

func (c client) GetPaymentPointer(ctx context.Context, ppURL string) (*openpayments.PaymentPointer, error) {
	return ops.GetPaymentPointer(ctx, c.b, ppURL)
}
