package linkedaccounts

import "context"

type Client interface {
	Create(ctx context.Context, args *CreateArgs) (*LinkedAccount, error)
	Get(ctx context.Context, id string) (*LinkedAccount, error)
	GetByProviderID(ctx context.Context, args GetByProviderIDArgs) (*LinkedAccount, error)
	ListByWalletId(ctx context.Context, walletId string) ([]LinkedAccount, error)
}
