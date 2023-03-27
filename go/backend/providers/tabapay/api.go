package tabapay

import (
	"context"
)

type Client interface {
	CreateCard(ctx context.Context, args CreateCardArgs) (Await, error)
}
