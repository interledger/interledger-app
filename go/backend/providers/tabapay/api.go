package tabapay

import (
	"context"

	"gitlab.com/fynbos/backend/linkedaccounts"
)

type Client interface {
	CreateCard(ctx context.Context, args CreateCardArgs) (*linkedaccounts.LinkedAccount, error)
}
