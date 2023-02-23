package openpayments

import (
	"context"
)

type Client interface {
	GetWalletPaymentPointer(ctx context.Context, walletID string) (*PaymentPointer, error)
	GetPaymentPointer(ctx context.Context, ppURL string) (*PaymentPointer, error)
}
