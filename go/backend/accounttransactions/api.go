package account_transactions

import (
	"context"

	"github.com/jmoiron/sqlx"
)

type Client interface {
	// TODO: Create should return an error array
	Create(ctx context.Context, args *CreateTransactionArgs) (*AccountTransaction, error)
	CreatePending(ctx context.Context, args *CreatePendingTransactionArgs) (*AccountTransaction, error)
	PostPending(ctx context.Context, id string) (*AccountTransaction, error)
	VoidPending(ctx context.Context, id string) (*AccountTransaction, error)
	GetByAccount(ctx context.Context, t *sqlx.Tx, args *GetByAccountArgs) ([]*AccountTransaction, error)
	GetPage(ctx context.Context, args *PaginationArgs) ([]AccountTransaction, error)
	GetPageInfo(
		ctx context.Context,
		accountID string,
		edges []AccountTransaction,
	) (
		hasNextPage bool,
		startCursor string,
		endCursor string,
		err error,
	)
	Get(ctx context.Context, id string) (*AccountTransaction, error)
}
