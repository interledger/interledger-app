package client

import (
	"context"

	"github.com/interledger/interledger-app/go/backend/identities"
	"github.com/interledger/interledger-app/go/backend/identities/ops"
)

var _ identities.Client = client{}

type client struct {
	b ops.Backends
}

func New(b ops.Backends) identities.Client {
	return &client{
		b: b,
	}
}

func (c client) List(ctx context.Context, walletID string) ([]identities.Identity, error) {
	return ops.List(ctx, c.b, walletID)
}

func (c client) ListPublic(ctx context.Context, walletID string) ([]identities.Identity, error) {
	return ops.ListPublic(ctx, c.b, walletID)
}

func (c client) Add(ctx context.Context, args identities.AddArgs) (*identities.Identity, error) {
	return ops.Add(ctx, c.b, args)
}

func (c client) VerifyInstructions(ctx context.Context, id string) (*identities.VerifyInstructions, error) {
	return ops.VerifyInstructions(ctx, c.b, id)
}

func (c client) StartVerification(ctx context.Context, id, proof string) (*identities.Identity, error) {
	return ops.StartVerification(ctx, c.b, id, proof)
}

func (c client) Delete(ctx context.Context, id, walletID string) error {
	return ops.Delete(ctx, c.b, id, walletID)
}

func (c client) SetPublic(ctx context.Context, id, walletID string, public bool) (*identities.Identity, error) {
	return ops.SetPublic(ctx, c.b, id, walletID, public)
}

func (c client) Get(ctx context.Context, id string) (*identities.Identity, error) {
	return ops.Get(ctx, c.b, id)
}

func (c client) UpdateState(ctx context.Context, id string, state identities.State, proof string) error {
	return ops.UpdateState(ctx, c.b, id, state, proof)
}

func (c client) GetBySignatureHash(ctx context.Context, sigHash []byte) (*identities.Identity, error) {
	return ops.GetBySignatureHash(ctx, c.b, sigHash)
}

func (c client) GetByIdentifier(ctx context.Context, identifier string) (*identities.Identity, error) {
	return ops.GetByIdentifier(ctx, c.b, identifier)
}

func (c client) Search(ctx context.Context, walletID, term string) ([]identities.SearchResult, error) {
	return ops.Search(ctx, c.b, walletID, term)
}
