package basistheory

import (
	"context"
)

type Client interface {
	CreateCard(ctx context.Context, tokenID, walletID string) (*Card, error)
	GetCard(ctx context.Context, id string) (*Card, error)
}
