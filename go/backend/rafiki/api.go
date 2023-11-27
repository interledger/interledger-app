package rafiki

import (
	"context"
	"net/http"

	"gitlab.com/fynbos/backend/wallets"
)

type Client interface {
	WebhookHandler() http.HandlerFunc
	CreateWalletAddress(ctx context.Context, address wallets.Wallet) error
	FundOutgoingPayment(ctx context.Context, paymentID string) error
}
