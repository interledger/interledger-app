package identities

import (
	"context"
)

type Client interface {
	List(ctx context.Context, walletID string) ([]Identity, error)
	ListPublic(ctx context.Context, walletID string) ([]Identity, error)
	Add(ctx context.Context, args AddArgs) (*VerifyInstructions, error)
	VerifyInstruction(ctx context.Context, id string) (*VerifyInstructions, error)
	Verify(ctx context.Context, id string) (*Identity, error)
	Delete(ctx context.Context, id string) error
	SetPublic(ctx context.Context, id string, public bool) (*Identity, error)
}
