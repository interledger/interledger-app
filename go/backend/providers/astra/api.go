package astra

import (
	"context"
	"net/http"
)

type Client interface {
	WebhookHandler() http.HandlerFunc
	StartKYC(ctx context.Context, walletID string) error
	CreateCard(ctx context.Context, args CreateCardArgs) (Await, error)
}
