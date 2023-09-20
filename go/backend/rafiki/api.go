package rafiki

import (
	"context"
	"net/http"

	"gitlab.com/fynbos/backend/wallets"
)

type Client interface {
	WebhookHandler() http.HandlerFunc
	CreatePaymentPointer(ctx context.Context, address wallets.Wallet) error
}
