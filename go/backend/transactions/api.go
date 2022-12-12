package transactions

import (
	"context"

	"github.com/jmoiron/sqlx"
)

type Client interface {
	CreateTransaction(ctx context.Context, args CreateTransactionArgs) error
	CreateTransactionTx(ctx context.Context, tx *sqlx.Tx, args CreateTransactionArgs) error
	UpdateTransaction(ctx context.Context, args UpdateTransactionArgs) error
	UpdateTransactionTx(ctx context.Context, tx *sqlx.Tx, args UpdateTransactionArgs) error
	AddTransfers(ctx context.Context, args []CreateTransferArgs) error
	AddTransfersTx(ctx context.Context, tx *sqlx.Tx, args []CreateTransferArgs) error
}
