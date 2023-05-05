package features

import (
	"context"
)

type Client interface {
	SetFeatures(ctx context.Context, walletID string, features WalletFeatures) (*WalletFeatures, error)
	Features(ctx context.Context, walletID string) (*WalletFeatures, error)
}
