package mx

import (
	"context"

	"gitlab.com/fynbos/backend/linkedaccounts"
)

type Client interface {
	GetWidget(ctx context.Context, walletID string) (string, error)
	CreateBankAccounts(ctx context.Context, args CreateBankAccountsArgs) ([]linkedaccounts.LinkedAccount, error)
	GetAccount(ctx context.Context, walletID, accountGuid string) (*Account, error)
	ListUsers(ctx context.Context) ([]User, error)
	DeleteExternalUser(ctx context.Context, guid string) error
}
