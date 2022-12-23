package client

import (
	"context"

	"gitlab.com/fynbos/backend/db"

	"github.com/jmoiron/sqlx"
	"gitlab.com/fynbos/backend/transactions"
	"gitlab.com/fynbos/backend/transactions/ops"
)

var _ transactions.Client = &client{}

type client struct {
	b ops.Backends
}

func New(b ops.Backends) transactions.Client {
	return &client{
		b: b,
	}
}

func (c *client) CreateTransaction(ctx context.Context, args transactions.CreateTransactionArgs) error {
	return ops.CreateTransaction(ctx, c.b, args)
}

func (c *client) CreateTransactionTx(ctx context.Context, tx *sqlx.Tx, args transactions.CreateTransactionArgs) error {
	return ops.CreateTransactionTx(ctx, c.b, tx, args)
}

func (c *client) UpdateTransaction(ctx context.Context, args transactions.UpdateTransactionArgs) error {
	return ops.UpdateTransaction(ctx, c.b, args)
}

func (c *client) UpdateTransactionTx(ctx context.Context, tx *sqlx.Tx, args transactions.UpdateTransactionArgs) error {
	return ops.UpdateTransactionTx(ctx, c.b, tx, args)
}

func (c *client) AddTransfers(ctx context.Context, args []transactions.TransferArgs) error {
	return ops.AddTransfers(ctx, c.b, args)
}

func (c *client) AddTransfersTx(ctx context.Context, tx *sqlx.Tx, args []transactions.TransferArgs) error {
	return ops.AddTransfersTx(ctx, c.b, tx, args)
}

func (c *client) UpdateTransfers(ctx context.Context, args []transactions.TransferArgs) error {
	return ops.UpdateTransfers(ctx, c.b, args)
}

func (c *client) UpdateTransfersTx(ctx context.Context, tx *sqlx.Tx, args []transactions.TransferArgs) error {
	return ops.UpdateTransfersTx(ctx, c.b, tx, args)
}

func (c *client) UpdateForeignIDs(ctx context.Context, args transactions.UpdateForeignIDArgs) error {
	return ops.UpdateForeignIDs(ctx, c.b, args)
}

func (c *client) ListTransactions(ctx context.Context, page db.Pagination, walletID string) ([]transactions.Transaction, error) {
	return ops.ListTransactions(ctx, c.b, walletID, page)
}

func (c *client) GetTransaction(ctx context.Context, walletID string, txID string) (*transactions.Transaction, error) {
	return ops.GetTransaction(ctx, c.b, walletID, txID)
}
