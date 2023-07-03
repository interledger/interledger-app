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

func (c client) GetOutgoingPayment(ctx context.Context, id string) (*openpayments.OutgoingPayment, error) {
	return ops.GetOutgoingPayment(ctx, c.b, id)
}

func (c client) GetIncomingPayment(ctx context.Context, id string) (*openpayments.IncomingPayment, error) {
	return ops.GetIncomingPayment(ctx, c.b, id)
}

func (c client) GetQuote(ctx context.Context, id string) (*openpayments.Quote, error) {
	return ops.GetQuote(ctx, c.b, id)
}
