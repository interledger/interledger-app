package identities

import (
	"context"
)

type Client interface {
	List(ctx context.Context, walletID string) ([]Identity, error)
	ListPublic(ctx context.Context, walletID string) ([]Identity, error)
	Add(ctx context.Context, args AddArgs) (*Identity, error)
	VerifyInstructions(ctx context.Context, id string) (*VerifyInstructions, error)
	StartVerification(ctx context.Context, id, proof string) (*Identity, error)
	Delete(ctx context.Context, id, walletID string) error
	SetPublic(ctx context.Context, id, walletID string, public bool) (*Identity, error)
	Get(ctx context.Context, id string) (*Identity, error)
	UpdateState(ctx context.Context, id string, state State, proof string) error
}

// 1. Create connection [x]
// 2. On callback Add identity [x]
// 3. Post tweet [x]
// 4. Start Verification
