package openpayments

import (
	"context"

	"gitlab.com/fynbos/backend/db"
)

type Client interface {
	ListTransactions(ctx context.Context, walletID string, pagination db.Pagination) ([]Transaction, error)
}
