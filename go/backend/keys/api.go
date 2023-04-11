package keys

import (
	"context"
)

type Client interface {
	ProvisionPrivateKey(ctx context.Context, walletID string) error
	AddPublicKey(ctx context.Context, walletID string, publicKeyBase64 string, name string) error
	List(ctx context.Context, walletID string) ([]Key, error)

	Verify(ctx context.Context, keyID string, walletID string, message, signature []byte) (bool, error)
	Sign(ctx context.Context, keyID string, walletID string, message []byte) ([]byte, error)
}
