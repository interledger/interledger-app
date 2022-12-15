package client

import (
	"context"

	"gitlab.com/fynbos/backend/db"
	"gitlab.com/fynbos/backend/openpayments"
	"gitlab.com/fynbos/backend/openpayments/ops"
)

var _ openpayments.Client = &client{}

type client struct {
	b ops.Backends
}

func New(b ops.Backends) openpayments.Client {
	return &client{
		b: b,
	}
}

func (c client) ListTransactions(ctx context.Context, walletID string, pagination db.Pagination) ([]openpayments.Transaction, error) {
	return nil, nil
}
