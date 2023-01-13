package statements

import (
	"context"

	"gitlab.com/fynbos/backend/providers/machnet"
	"gitlab.com/fynbos/backend/transactions"
)

type Client interface {
	GenerateWalletStatementPDF(ctx context.Context, wallet *machnet.Wallet, transactions []transactions.Transaction) ([]byte, error)
}
