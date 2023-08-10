package payments

import "context"

type Client interface {
	Create(ctx context.Context, payment Payment) (*Payment, error)
	Update(ctx context.Context, payment Payment) (*Payment, error)
	Confirm(ctx context.Context, paymentID string) (*Payment, error)
}
