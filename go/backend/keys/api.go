package keys

import (
	"context"
)

type Client interface {
	GetPublicKey(ctx context.Context, id string, walletID string) (*Key, error)
	ProvisionPrivateKey(ctx context.Context, walletID string) error
	AddPublicKey(ctx context.Context, walletID, publicKeyBase64, name, keyID string) (*Key, error)
	DeletePublicKey(ctx context.Context, id string) error
	List(ctx context.Context, walletID string) ([]Key, error)
	RemoveCustodialKeysForWallet(ctx context.Context, walletID string) error

	Verify(ctx context.Context, keyID, walletID string, message, signature []byte) (bool, error)
	Sign(ctx context.Context, keyID, walletID string, message []byte) ([]byte, error)

	FixWalletPublicKey(ctx context.Context, walletID string) error
}
