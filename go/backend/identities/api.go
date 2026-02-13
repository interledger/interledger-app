package identities

import (
	"context"
)

type Client interface {
	List(ctx context.Context, walletID string) ([]Identity, error)
	ListPublic(ctx context.Context, walletID string) ([]Identity, error)
	VerifyInstructions(ctx context.Context, id string) (*VerifyInstructions, error)
	StartVerification(ctx context.Context, id, proof string) (*Identity, error)
	Delete(ctx context.Context, id, walletID string) error
	SetPublic(ctx context.Context, id, walletID string, public bool) (*Identity, error)
	Get(ctx context.Context, id string) (*Identity, error)
	UpdateState(ctx context.Context, id string, state State, proof string) error
	GetBySignatureHash(ctx context.Context, sigHash []byte) (*Identity, error)
	GetByIdentifier(ctx context.Context, identifier string) (*Identity, error)
	Search(ctx context.Context, walletID, term string) ([]SearchResult, error)
}
