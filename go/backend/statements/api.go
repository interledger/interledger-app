package statements

import (
	"context"

	"gitlab.com/fynbos/backend/db"
	"gitlab.com/fynbos/backend/providers/machnet"
	"gitlab.com/fynbos/backend/transactions"
)

type Client interface {
	GenerateWalletStatementPDF(ctx context.Context, wallet *machnet.Wallet, transactions []transactions.Transaction) ([]byte, error)
	Store(ctx context.Context, statement Statement) error
	List(ctx context.Context, page db.Pagination) ([]Statement, error)
	GetSignedURL(ctx context.Context, id string) (*Statement, error)
}
