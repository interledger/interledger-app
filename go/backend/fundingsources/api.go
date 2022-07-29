package fundingsources

import "context"

type Client interface {
	Create(ctx context.Context, args *CreateArgs) (*FundingSource, error)
	Get(ctx context.Context, id string) (*FundingSource, error)
	GetByAccountId(ctx context.Context, identityId string) ([]FundingSource, error)
	Verify(ctx context.Context, args *VerifyArgs) (*FundingSource, error)
	CreateBankAccount(ctx context.Context, args *CreateBankAccountArgs) (*FundingSource, error)
}
