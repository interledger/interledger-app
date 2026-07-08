package client

import (
	"context"
	"time"

	"github.com/interledger/interledger-app/go/backend/currency"

	"github.com/interledger/interledger-app/go/backend/db"
	"github.com/interledger/interledger-app/go/backend/linkedaccounts"
	"github.com/interledger/interledger-app/go/backend/linkedaccounts/ops"
	"github.com/interledger/interledger-app/go/log"
	"go.uber.org/zap"
)

var _ linkedaccounts.Client = client{}

type client struct {
	b ops.Backends
}

func New(b ops.Backends) linkedaccounts.Client {
	return &client{
		b: b,
	}
}

func (c client) ListBalances(ctx context.Context, walletID string) ([]linkedaccounts.LinkedAccount, error) {
	return ops.ListBalances(ctx, c.b, walletID)
}

func (c client) Create(ctx context.Context, args *linkedaccounts.CreateArgs) (fs *linkedaccounts.LinkedAccount, err error) {
	defer func(begin time.Time) {
		if err != nil {
			log.Error(
				"Failed to create linked account.",
				zap.Int64("took", time.Since(begin).Milliseconds()),
				zap.String("msg", err.Error()),
			)
			return
		}

		log.Debug(
			"Created linked account.",
			zap.Int64("took", time.Since(begin).Milliseconds()),
		)
	}(time.Now())

	return ops.Create(ctx, c.b, args)
}

func (c client) CreateBatch(ctx context.Context, args []linkedaccounts.CreateArgs) ([]linkedaccounts.LinkedAccount, error) {
	return ops.CreateBatch(ctx, c.b, args)
}

func (c client) Get(ctx context.Context, id string) (fs *linkedaccounts.LinkedAccount, err error) {
	defer func(begin time.Time) {
		if err != nil {
			log.Error(
				"Failed to get linked account.",
				zap.String("id", id),
				zap.Int64("took", time.Since(begin).Milliseconds()),
				zap.String("msg", err.Error()),
			)
			return
		}

		log.Debug(
			"Got linked account.",
			zap.String("id", fs.ID),
			zap.Int64("took", time.Since(begin).Milliseconds()),
		)
	}(time.Now())

	return ops.Get(ctx, c.b, id)
}

func (c client) GetByProviderID(ctx context.Context, args linkedaccounts.GetByProviderIDArgs) (fs *linkedaccounts.LinkedAccount, err error) {
	return ops.GetByProviderID(ctx, c.b, args)
}

func (c client) ListByWalletId(ctx context.Context, walletId string) (fsl []linkedaccounts.LinkedAccount, err error) {
	defer func(begin time.Time) {
		if err != nil {
			log.Error(
				"Failed to get linked accounts.",
				zap.String("walletId", walletId),
				zap.Int64("took", time.Since(begin).Milliseconds()),
				zap.String("msg", err.Error()),
			)
			return
		}

		log.Debug(
			"Got linked accounts.",
			// zap.String("id", fs[0]),
			zap.Int64("took", time.Since(begin).Milliseconds()),
		)
	}(time.Now())

	return ops.ListByWalletId(ctx, c.b, walletId)
}

func (c client) GetDefaultReceive(ctx context.Context, walletID string, cc currency.Currency) (*linkedaccounts.LinkedAccount, error) {
	return ops.GetDefaultReceive(ctx, c.b, walletID, cc)
}

func (c client) GetDefaultSend(ctx context.Context, walletID string, cc currency.Currency) (*linkedaccounts.LinkedAccount, error) {
	return ops.GetDefaultSend(ctx, c.b, walletID, cc)
}

func (c client) Delete(ctx context.Context, id string) error {
	return ops.Delete(ctx, c.b, id)
}

func (c client) MarkNotDeleted(ctx context.Context, id string) (*linkedaccounts.LinkedAccount, error) {
	return ops.MarkNotDeleted(ctx, c.b, id)
}

func (c client) ListByProviderID(ctx context.Context, provider, providerID string) ([]linkedaccounts.LinkedAccount, error) {
	return ops.ListByProviderID(ctx, c.b, provider, providerID)
}

func (c client) SetNickname(ctx context.Context, id, nickname string) (*linkedaccounts.LinkedAccount, error) {
	err := ops.SetNickname(ctx, c.b, id, nickname)
	if err != nil {
		return nil, err
	}

	return ops.Get(ctx, c.b, id)
}

func (c client) CreateReviews(ctx context.Context, args []linkedaccounts.CreateReviewArgs) ([]linkedaccounts.Review, error) {
	return ops.CreateReviews(ctx, c.b, args)
}

func (c client) GetReview(ctx context.Context, id string) (*linkedaccounts.Review, error) {
	return ops.GetReview(ctx, c.b, id)
}

func (c client) ListReviews(ctx context.Context, pagination db.Pagination) ([]linkedaccounts.Review, error) {
	return ops.ListReviews(ctx, c.b, pagination)
}

func (c client) ListIncompleteReviews(ctx context.Context, pagination db.Pagination) ([]linkedaccounts.Review, error) {
	return ops.ListIncompleteReviews(ctx, c.b, pagination)
}

func (c client) CompleteReview(ctx context.Context, args linkedaccounts.CompleteReviewArgs) (*linkedaccounts.Review, error) {
	return ops.CompleteReview(ctx, c.b, args)
}

func (c client) SetDefaultSend(ctx context.Context, id string) (*linkedaccounts.LinkedAccount, error) {
	return ops.SetDefaultSend(ctx, c.b, id)
}

func (c client) SetDefaultReceive(ctx context.Context, id string) (*linkedaccounts.LinkedAccount, error) {
	return ops.SetDefaultReceive(ctx, c.b, id)
}

func (c client) CanSendToWallet(ctx context.Context, sendWalletID, recvWalletID string) (bool, error) {
	return ops.CanSendToWallet(ctx, c.b, sendWalletID, recvWalletID)
}
