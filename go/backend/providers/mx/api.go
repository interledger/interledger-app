package mx

import (
	"context"

	"gitlab.com/fynbos/backend/linkedaccounts"
)

type Client interface {
	GetWidget(ctx context.Context, walletID string) (string, error)
	CreateBankAccounts(ctx context.Context, args CreateBankAccountsArgs) ([]linkedaccounts.LinkedAccount, error)
}
