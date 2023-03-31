package keys

import (
	"context"
)

type Client interface {
	ProvisionPrivateKey(ctx context.Context, walletID string) error
	AddPublicKey(ctx context.Context, walletID string, publicKeyBase64 string, name string) error
	List(ctx context.Context, walletID string) ([]Key, error)

	// Should these require walletID as well?
	Verify(ctx context.Context, keyID string, message, signature []byte) (bool, error)
	Sign(ctx context.Context, keyID string, message []byte) ([]byte, error)
}
