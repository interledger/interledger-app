package client

import (
	"context"
	"gitlab.com/fynbos/cli/identities"
	"gitlab.com/fynbos/cli/identities/ops"
)

var _ identities.Client = client{}

type client struct {
}

func New() identities.Client {
	return client{}
}

func (c client) Verify(ctx context.Context, args *identities.VerifyArgs) error {
	err := ops.VerifyIdentity(ctx, args)
	if err != nil {
		return err
	}

	return nil
}
