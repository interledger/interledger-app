package features

import "context"

type Client interface {
	Features(ctx context.Context, walletID string)
}

type WalletFeatures struct {
	SendEnabled       bool
	RecvEnabled       bool
	LinkedAccEnabled  bool
	CardsEnabled      bool
	BanksEnabled      bool
	IdentitiesEnabled bool
	TwitterEnabled    bool
}
