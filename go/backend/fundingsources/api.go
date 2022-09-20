package fundingsources

import "context"

type Client interface {
	Create(ctx context.Context, args *CreateArgs) (*FundingSource, error)
	Get(ctx context.Context, id string) (*FundingSource, error)
	GetByWalletId(ctx context.Context, walletId string) ([]FundingSource, error)
	Verify(ctx context.Context, args *VerifyArgs) (*FundingSource, error)
}
