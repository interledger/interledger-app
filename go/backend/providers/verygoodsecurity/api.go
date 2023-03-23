package verygoodsecurity

import (
	"context"
)

type Client interface {
	CreateCard(ctx context.Context, args Card) (*Card, error)
}
