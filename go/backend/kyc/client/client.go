package client

import (
	"context"

	"gitlab.com/fynbos/backend/kyc"
	"gitlab.com/fynbos/backend/kyc/ops"
)

var _ kyc.Client = client{}

type client struct {
	b ops.Backends
}

func New(b ops.Backends) kyc.Client {
	return &client{
		b: b,
	}
}

func (c client) GetUserDetails(ctx context.Context, userID string) (*kyc.UserDetails, error) {
	return ops.GetUserDetails(ctx, c.b, userID)
}

func (c client) UpdateUserDetails(ctx context.Context, args kyc.UserDetails) (*kyc.UserDetails, error) {
	return ops.UpdateUserDetails(ctx, c.b, args)
}
