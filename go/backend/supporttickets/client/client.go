package client

import (
	"context"

	"gitlab.com/fynbos/backend/supporttickets"
	"gitlab.com/fynbos/backend/supporttickets/ops"
	"gitlab.com/fynbos/backend/supporttickets/zendesk"
)

var _ supporttickets.Client = client{}

type client struct {
	b  ops.Backends
	zc zendesk.Client
}

func NewClient(b ops.Backends, zendeskToken string) supporttickets.Client {
	// Inject the zendesk client that the rest of the backends services don't need or use.
	return client{b: b, zc: zendesk.NewClient(zendeskToken)}
}

func (c client) CreateTicket(ctx context.Context, args supporttickets.CreateTicketArgs) error {
	return ops.CreateTicket(ctx, c.b, c.zc, args)
}
