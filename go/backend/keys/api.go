package keys

import (
	"context"
)

type Client interface {
	ProvisionPrivateKey(ctx context.Context, walletID string) error
	AddPublicKey(ctx context.Context, walletID string, publicKeyBase64 string, name string) error
}
