package client

import (
	"context"

	"github.com/interledger/interledger-app/go/backend/signup"

	"github.com/interledger/interledger-app/go/backend/signup/ops"
)

var _ signup.Client = client{}

type client struct {
	b ops.Backends
}

func New(b ops.Backends) signup.Client {
	return &client{
		b: b,
	}
}

func (c client) Get(ctx context.Context, id string) (*signup.Signup, error) {
	return ops.Get(ctx, c.b, id)
}

func (c client) SetUserData(ctx context.Context, args signup.UserDataArgs) (string, error) {
	return ops.SetUserData(ctx, c.b, args)
}

func (c client) SetMobileNumber(ctx context.Context, args signup.MobileNumberArgs) error {
	return ops.SetMobileNumber(ctx, c.b, args)
}

func (c client) Complete(ctx context.Context, id, userID string) error {
	return ops.Complete(ctx, c.b, id, userID)
}

func (c client) GetForUser(ctx context.Context, userID string) (*signup.Signup, error) {
	return ops.GetForUser(ctx, c.b, userID)
}
