package linkedaccounts

import (
	"context"

	"github.com/interledger/interledger-app/go/backend/currency"

	"github.com/interledger/interledger-app/go/backend/db"
)

type Client interface {
	Create(ctx context.Context, args *CreateArgs) (*LinkedAccount, error)
	CreateBatch(ctx context.Context, args []CreateArgs) ([]LinkedAccount, error)
	Get(ctx context.Context, id string) (*LinkedAccount, error)
	Delete(ctx context.Context, id string) error
	MarkNotDeleted(ctx context.Context, id string) (*LinkedAccount, error)
	GetByProviderID(ctx context.Context, args GetByProviderIDArgs) (*LinkedAccount, error)
	ListByWalletId(ctx context.Context, walletID string) ([]LinkedAccount, error)
	ListByProviderID(ctx context.Context, provider, providerID string) ([]LinkedAccount, error)
	SetNickname(ctx context.Context, id, nickname string) (*LinkedAccount, error)
	ListBalances(ctx context.Context, walletID string) ([]LinkedAccount, error)

	GetDefaultReceive(ctx context.Context, walletID string, cc currency.Currency) (*LinkedAccount, error)
	GetDefaultSend(ctx context.Context, walletID string, cc currency.Currency) (*LinkedAccount, error)
	SetDefaultReceive(ctx context.Context, id string) (*LinkedAccount, error)
	SetDefaultSend(ctx context.Context, id string) (*LinkedAccount, error)
	CanSendToWallet(ctx context.Context, sendWalletID, receiveWalletID string) (bool, error)

	CreateReviews(ctx context.Context, args []CreateReviewArgs) ([]Review, error)
	GetReview(ctx context.Context, id string) (*Review, error)
	ListReviews(ctx context.Context, pagination db.Pagination) ([]Review, error)
	ListIncompleteReviews(ctx context.Context, pagination db.Pagination) ([]Review, error)
	CompleteReview(ctx context.Context, args CompleteReviewArgs) (*Review, error)
}
