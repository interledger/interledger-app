package client

import (
	"context"

	"github.com/interledger/interledger-app/go/backend/agreements"
	"github.com/interledger/interledger-app/go/backend/agreements/ops"
)

var _ agreements.Client = client{}

type client struct {
	b ops.Backends
}

func New(b ops.Backends) agreements.Client {
	return &client{
		b: b,
	}
}

func (c client) Sign(ctx context.Context, args *agreements.SignArgs) error {
	return ops.Sign(ctx, c.b, args)
}

func (c client) GetSignatures(ctx context.Context, userID string) ([]agreements.Signature, error) {
	return ops.GetSignatures(ctx, c.b, userID)
}

func (c client) Get(ctx context.Context, id string) (*agreements.Agreement, error) {
	return ops.Get(ctx, c.b, id)
}

func (c client) ListAffectedUserIDsPaginated(ctx context.Context, changes []agreements.AgreementChange, limit, offset int) ([]string, error) {
	return ops.ListAffectedUserIDsPaginated(ctx, c.b, changes, limit, offset)
}

func (c client) GetAgreementNamesSignedByUsersFromSet(ctx context.Context, userIDs []string, changes []agreements.AgreementChange) (map[string][]string, error) {
	return ops.GetAgreementNamesSignedByUsersFromSet(ctx, c.b, userIDs, changes)
}

func (c client) MarkUsersNotified(ctx context.Context, userIDs []string, changes []agreements.AgreementChange) error {
	return ops.MarkUsersNotified(ctx, c.b, userIDs, changes)
}
