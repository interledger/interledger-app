package xago

import (
	"context"
	"net/http"
)

type Client interface {
	WebhookHandler() http.HandlerFunc
	CreateSubAccount(ctx context.Context, walletID string) (Await, error)
	CreateBeneficiary(ctx context.Context, walletID string) (Await, error)
}
