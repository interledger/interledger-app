package keys

import (
	"context"
)

type Client interface {
	GetPublicKey(ctx context.Context, id string, walletID string) (*Key, error)
	AddPublicKey(ctx context.Context, walletID, publicKeyBase64, name, keyID string) (*Key, error)
	DeletePublicKey(ctx context.Context, id string) error
	List(ctx context.Context, walletID string) ([]Key, error)
	RemoveCustodialKeysForWallet(ctx context.Context, walletID string) error
}
