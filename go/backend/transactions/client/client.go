package client

import (
	"context"

	"github.com/jmoiron/sqlx"
	"gitlab.com/fynbos/backend/kyc/address"
	"gitlab.com/fynbos/backend/transactions"
	"gitlab.com/fynbos/backend/transactions/ops"
)

var _ transactions.Client = &client{}

type client struct {
	b   ops.Backends
	val address.Validator
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

func (c *client) AddTransfer(ctx context.Context, args transactions.CreateTransferArgs) error {
	return ops.AddTransfer(ctx, c.b, args)
}

func (c *client) AddTransferTx(ctx context.Context, tx *sqlx.Tx, args transactions.CreateTransferArgs) error {
	return ops.AddTransferTx(ctx, c.b, tx, args)
}
