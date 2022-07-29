package identity

import "context"

type Client interface {
	Create(ctx context.Context, args *CreateArgs) (*Identity, error)
	Get(ctx context.Context, id string) (*Identity, error)
	GetByEmail(ctx context.Context, email string) (*Identity, error)
}
