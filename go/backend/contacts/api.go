package contacts

import (
	"context"
	"gitlab.com/fynbos/backend/db"
)

type Client interface {
	Create(ctx context.Context, args CreateContactArgs) (*Contact, error)
	List(ctx context.Context, walletID string, page db.Pagination) ([]Contact, error)
	Get(ctx context.Context, walletID string, pp string) (*Contact, error)
}
