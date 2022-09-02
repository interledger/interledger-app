package payments

import "context"

type Client interface {
	InitiateOutgoingPayment(ctx context.Context, args InitiateOutgoingPaymentArgs) (*OutgoingPayment, error)
	Get(ctx context.Context, id string) (*OutgoingPayment, error)
	SetState(ctx context.Context, id string, state State) error
}
