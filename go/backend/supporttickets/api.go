package supporttickets

import "context"

type Client interface {
	CreateTicket(ctx context.Context, args CreateTicketArgs) error
}
