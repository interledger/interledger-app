package openpayments

import (
	"context"
)

type Client interface {
	GetOutgoingPayment(ctx context.Context, id string) (*OutgoingPayment, error)
	GetIncomingPayment(ctx context.Context, id string) (*IncomingPayment, error)
	GetQuote(ctx context.Context, id string) (*Quote, error)
}
