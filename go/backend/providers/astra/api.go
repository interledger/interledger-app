package astra

import (
	"context"
	"net/http"
)

type Client interface {
	WebhookHandler() http.HandlerFunc
	TrustedAuthInfoWebhook() http.HandlerFunc
	StartKYC(ctx context.Context, walletID string) error
	CreateCard(ctx context.Context, args CreateCardArgs) (Await, error)
	DebitCard(ctx context.Context, args CardToAccountArgs) (string, error)
	CreditCard(ctx context.Context, args AccountToCardsArgs) (string, error)
	LookupTransfer(ctx context.Context, walletID, txID string) (*Transfer, error)
	LookupRoutine(ctx context.Context, walletID, routineID string) (*Routine, error)
}
