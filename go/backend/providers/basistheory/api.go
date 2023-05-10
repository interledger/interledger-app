package basistheory

import (
	"context"
)

type Client interface {
	CreateCard(ctx context.Context, tokenID, walletID string) (*Card, error)
	CreateCardToken(ctx context.Context, args CreateCardArgs) (string, error)
	GetCard(ctx context.Context, id string) (*Card, error)
}
