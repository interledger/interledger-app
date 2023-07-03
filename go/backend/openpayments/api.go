package openpayments

import (
	"context"
)

type Client interface {
	GetWalletPaymentPointer(ctx context.Context, walletID string) (*PaymentPointer, error)
	GetPaymentPointer(ctx context.Context, ppURL string) (*PaymentPointer, error)
	GetOutgoingPayment(ctx context.Context, id string) (*OutgoingPayment, error)
	GetIncomingPayment(ctx context.Context, id string) (*IncomingPayment, error)
	GetQuote(ctx context.Context, id string) (*Quote, error)
}
