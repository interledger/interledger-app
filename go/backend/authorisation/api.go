package authorisation

import "context"

type InternalClient interface {
	AddPublicKey(ctx context.Context, walletID string, key string) error
	Introspect(ctx context.Context, token string) error
	ListKeys(ctx context.Context, paymentPointer string) ([]string, error)
}
