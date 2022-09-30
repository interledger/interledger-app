package openpayments

import "context"

type Client interface {
	CreatePaymentPointer(ctx context.Context, pointer PaymentPointer) (*PaymentPointer, error)
	GetPaymentPointer(ctx context.Context, url string) (*PaymentPointer, error)
	PaymentPointerExists(ctx context.Context, url string) (bool, error)
	ListWalletPaymentPointers(ctx context.Context, walletID string) ([]PaymentPointer, error)
}
