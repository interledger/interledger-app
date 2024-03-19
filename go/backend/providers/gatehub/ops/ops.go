package ops

import (
	"context"
	"fmt"

	"gitlab.com/fynbos/backend/providers/gatehub"
	"gitlab.com/fynbos/backend/providers/gatehub/external"
)

func CreateUser(ctx context.Context, b Backends, ec external.Client, walletID string) error {
	ul, err := b.Users().ListUsers(ctx, walletID)
	if err != nil {
		return fmt.Errorf("%w %s", gatehub.ErrInternal, err)
	}
	if len(ul) < 1 {
		return fmt.Errorf("%w No Fynbos user found for walletID", gatehub.ErrInternal)
	}

	u, err := b.Users().GetUser(ctx, ul[0].ID)
	if err != nil {
		return fmt.Errorf("%w %s", gatehub.ErrInternal, err)
	}

	return ec.CreateUser(ctx, u.Email)
}
