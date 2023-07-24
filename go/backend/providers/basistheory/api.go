package basistheory

import (
	"context"
)

type Client interface {
	CreateCard(ctx context.Context, args CreateCardArgs) (*Card, error)
	CreateCardToken(ctx context.Context, args CreateCardTokenArgs) (string, error)
	GetCard(ctx context.Context, id string) (*Card, error)
}
