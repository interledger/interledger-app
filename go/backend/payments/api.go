package payments

import "context"

type Client interface {
	Lookup(ctx context.Context, id string) (*Payment, error)
	Create(ctx context.Context, args CreateArgs) (*Payment, error)
	Update(ctx context.Context, payment Payment) (*Payment, error)
	Confirm(ctx context.Context, id string) (*Payment, []RequiredActionType, error)
}
