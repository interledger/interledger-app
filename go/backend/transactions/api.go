package transactions

import (
	"context"

	"gitlab.com/fynbos/backend/db"

	"github.com/jmoiron/sqlx"
)

type Client interface {
	CreateTransaction(ctx context.Context, args CreateTransactionArgs) error
	CreateTransactionTx(ctx context.Context, tx *sqlx.Tx, args CreateTransactionArgs) error
	UpdateTransaction(ctx context.Context, args UpdateTransactionArgs) error
	UpdateTransactionTx(ctx context.Context, tx *sqlx.Tx, args UpdateTransactionArgs) error
	AddTransfers(ctx context.Context, args []TransferArgs) error
	AddTransfersTx(ctx context.Context, tx *sqlx.Tx, args []TransferArgs) error
	UpdateTransfers(ctx context.Context, args []TransferArgs) error
	UpdateTransfersTx(ctx context.Context, tx *sqlx.Tx, args []TransferArgs) error

	ListTransactions(ctx context.Context, page db.Pagination, walletID string) ([]Transaction, error)
	ListTransactionsInRange(ctx context.Context, walletID string, inRange TransactionRangeFilter) ([]Transaction, error)
	GetTransaction(ctx context.Context, walletID string, transactionID string) (*Transaction, error)
}
