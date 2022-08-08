package zendesk

import "context"

type Client interface {
	CreateTicket(ctx context.Context, email, name, description string) error
}
