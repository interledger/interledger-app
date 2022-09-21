package fundingsources

import "context"

type Client interface {
	Create(ctx context.Context, args *CreateArgs) (*FundingSource, error)
	Get(ctx context.Context, id string) (*FundingSource, error)
	ListByWalletId(ctx context.Context, walletId string) ([]FundingSource, error)
}
