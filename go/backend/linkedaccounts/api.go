package linkedaccounts

import (
	"context"

	"gitlab.com/fynbos/backend/db"
)

type Client interface {
	Create(ctx context.Context, args *CreateArgs) (*LinkedAccount, error)
	CreateBatch(ctx context.Context, args []CreateArgs) ([]LinkedAccount, error)
	Get(ctx context.Context, id string) (*LinkedAccount, error)
	Delete(ctx context.Context, id string) error
	MarkNotDeleted(ctx context.Context, id string) (*LinkedAccount, error)
	GetByProviderID(ctx context.Context, args GetByProviderIDArgs) (*LinkedAccount, error)
	ListByWalletId(ctx context.Context, walletId string) ([]LinkedAccount, error)
	ListMXBankAccounts(ctx context.Context) ([]LinkedAccount, error)
	SetNickname(ctx context.Context, id, nickname string) (*LinkedAccount, error)
	Requires3DS(ctx context.Context, id string) (bool, error)

	CreateReviews(ctx context.Context, args []CreateReviewArgs) ([]Review, error)
	GetReview(ctx context.Context, id string) (*Review, error)
	ListReviews(ctx context.Context, pagination db.Pagination) ([]Review, error)
	ListIncompleteReviews(ctx context.Context, pagination db.Pagination) ([]Review, error)
	CompleteReview(ctx context.Context, args CompleteReviewArgs) (*Review, error)
}
