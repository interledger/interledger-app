package authorisation

import "context"

type Client interface {
	AddPublicKey(ctx context.Context, walletID string, key string) error
	Introspect(ctx context.Context, token string) error
	ListKeys(ctx context.Context, paymentPointer string) ([]string, error)
}
