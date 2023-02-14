package linkedaccounts

import "context"

type Client interface {
	Create(ctx context.Context, args *CreateArgs) (*LinkedAccount, error)
	Get(ctx context.Context, id string) (*LinkedAccount, error)
	Delete(ctx context.Context, id string) error
	GetByProviderID(ctx context.Context, args GetByProviderIDArgs) (*LinkedAccount, error)
	ListByWalletId(ctx context.Context, walletId string) ([]LinkedAccount, error)
	ListMachnetWallets(ctx context.Context) ([]LinkedAccount, error)
	SetName(ctx context.Context, id, name string) (*LinkedAccount, error)
}
