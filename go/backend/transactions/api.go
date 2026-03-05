package transactions

import (
	"context"

	"gitlab.com/fynbos/backend/currency"

	"gitlab.com/fynbos/backend/db"

	"github.com/jmoiron/sqlx"
)

type Client interface {
	CreateTransaction(ctx context.Context, args CreateTransactionArgs) (string, error)
	CreateTransactionTx(ctx context.Context, tx *sqlx.Tx, args CreateTransactionArgs) (string, error)

	AddTransfers(ctx context.Context, trxID string, args []TransferArgs) error
	AddTransfersTx(ctx context.Context, tx *sqlx.Tx, trxID string, args []TransferArgs) error

	SetTransactionForeignID(ctx context.Context, ID string, foreignID string) error
	SetTransferForeignID(ctx context.Context, ID string, foreignID string) error
	SetTransactionDestination(ctx context.Context, id, destination string) error
	SetTransactionRefundState(ctx context.Context, ID string, state RefundState) error

	SetTransactionState(ctx context.Context, ID string, state State) error
	SetTransactionStateTx(ctx context.Context, tx *sqlx.Tx, ID string, state State) error
	SetTransferState(ctx context.Context, ID string, state State) error

	SetTransactionAmountTx(ctx context.Context, tx *sqlx.Tx, ID string, amount currency.Amount) error

	SetTransactionFee(ctx context.Context, ID string, fee currency.Amount) error

	GetHasTransacted(ctx context.Context, walletID, destination string) (bool, error)
	GetTransactedCount(ctx context.Context, walletID, destination string) (int, error)
	CountSendTransactions(ctx context.Context, walletID string) (int, error)
	ListTransactionsInRange(ctx context.Context, walletID string, inRange TransactionRangeFilter) ([]Transaction, error)
	List(ctx context.Context, page db.Pagination, walletID string) ([]Transaction, error)
	ListWithPending(ctx context.Context, page db.Pagination, walletID string) ([]Transaction, error)
	ListCompleted(ctx context.Context, page db.Pagination, walletID string) ([]Transaction, error)
	ListAll(ctx context.Context, page db.Pagination) ([]Transaction, error)
	GetTransaction(ctx context.Context, walletID string, trxID string) (*Transaction, error)
	GetTransactionByForeignID(ctx context.Context, walletID string, foreignID string) (*Transaction, error)
	ListTransfers(ctx context.Context, trxID string) ([]Transfer, error)
}
