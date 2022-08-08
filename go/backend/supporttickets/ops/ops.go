package ops

import (
	"context"
	"fmt"
	"strings"

	"gitlab.com/fynbos/backend/supporttickets/zendesk"

	"gitlab.com/fynbos/backend/supporttickets"
)

func CreateTicket(ctx context.Context, b Backends, zc zendesk.Client, args supporttickets.CreateTicketArgs) error {
	err := b.Validator().Struct(args)
	if err != nil {
		return err
	}

	return zc.CreateTicket(ctx, args.Email, strings.TrimSpace(fmt.Sprintf("%s %s", args.FirstName, args.LastName)), args.Description)
}
