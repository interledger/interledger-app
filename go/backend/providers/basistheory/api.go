package basistheory

import (
	"context"
)

type Client interface {
	CreateCard(ctx context.Context, args CreateCardArgs) (*Card, error)
	CreateCardToken(ctx context.Context, args CreateCardTokenArgs) (string, error)
	GetCard(ctx context.Context, id string) (*Card, error)
	GetCardByLinkedAccountID(ctx context.Context, linkedAccountID string) (*Card, error)
	UpdateCard(ctx context.Context, args UpdateCardArgs) (*Card, error)
	ListCards(ctx context.Context) ([]Card, error)
}
