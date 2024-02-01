package rafiki

import (
	"context"
	"net/http"

	"gitlab.com/fynbos/backend/wallets"
)

type Client interface {
	WebhookHandler() http.HandlerFunc
	CreatePaymentPointer(ctx context.Context, address wallets.Wallet, assetCode string) error
	CreatePaymentPointerKey(ctx context.Context, keyID string, walletID string) error
	RevokePaymentPointerKey(ctx context.Context, keyID string) error
	FundOutgoingPayment(ctx context.Context, paymentID string) error
	ListGrants(ctx context.Context, walletID string) ([]Grant, error)
	GetGrant(ctx context.Context, grantID string) (*Grant, error)
	RevokeGrant(ctx context.Context, grantID string) error
}
