package transactions

import (
	"context"

	"gitlab.com/fynbos/backend/db"

	"github.com/jmoiron/sqlx"
)

type Client interface {
	CreateTransaction(ctx context.Context, args CreateTransactionArgs) (string, error)
	CreateTransactionTx(ctx context.Context, tx *sqlx.Tx, args CreateTransactionArgs) (string, error)
	UpdateTransaction(ctx context.Context, args UpdateTransactionArgs) error
	UpdateTransactionTx(ctx context.Context, tx *sqlx.Tx, args UpdateTransactionArgs) error
	AddTransfers(ctx context.Context, args []TransferArgs) error
	AddTransfersTx(ctx context.Context, tx *sqlx.Tx, args []TransferArgs) error
	UpdateTransfers(ctx context.Context, args []TransferArgs) error
	UpdateTransfersTx(ctx context.Context, tx *sqlx.Tx, args []TransferArgs) error
	UpdateForeignIDs(ctx context.Context, args UpdateForeignIDArgs) error

	SetTransactionForeignID(ctx context.Context, ID string, foreignID string) error

	ListTransactions(ctx context.Context, page db.Pagination, walletID string) ([]Transaction, error)
	GetTransaction(ctx context.Context, walletID string, transactionID string) (*Transaction, error)
}
