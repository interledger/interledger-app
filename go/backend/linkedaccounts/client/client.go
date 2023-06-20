package client

import (
	"context"
	"time"

	"gitlab.com/fynbos/backend/linkedaccounts"
	"gitlab.com/fynbos/backend/linkedaccounts/ops"
	"gitlab.com/fynbos/log"
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

func (c client) Delete(ctx context.Context, id string) error {
	return ops.Delete(ctx, c.b, id)
}

func (c client) MarkNotDeleted(ctx context.Context, id string) (*linkedaccounts.LinkedAccount, error) {
	return ops.MarkNotDeleted(ctx, c.b, id)
}

func (c client) ListMXBankAccounts(ctx context.Context) ([]linkedaccounts.LinkedAccount, error) {
	return ops.ListMXBankAccounts(ctx, c.b)
}

func (c client) SetNickname(ctx context.Context, id, nickname string) (*linkedaccounts.LinkedAccount, error) {
	err := ops.SetNickname(ctx, c.b, id, nickname)
	if err != nil {
		return nil, err
	}

	return ops.Get(ctx, c.b, id)
}

func (c client) Requires3DS(ctx context.Context, id string) (bool, error) {
	return ops.Requires3DS(ctx, c.b, id)
}

func (c client) CreateReviews(ctx context.Context, args []linkedaccounts.CreateReviewArgs) ([]linkedaccounts.Review, error) {
	return ops.CreateReviews(ctx, c.b, args)
}

func (c client) GetReview(ctx context.Context, id string) (*linkedaccounts.Review, error) {
	return ops.GetReview(ctx, c.b, id)
}

func (c client) UpdateReviewState(ctx context.Context, id string, newState linkedaccounts.State) (*linkedaccounts.Review, error) {
	return ops.UpdateReviewState(ctx, c.b, id, newState)
}

func (c client) UpdateReviewReason(ctx context.Context, id, reason string) (*linkedaccounts.Review, error) {
	return ops.UpdateReviewReason(ctx, c.b, id, reason)
}

func (c client) CompleteReview(ctx context.Context, id, reviewedBy string) (*linkedaccounts.Review, error) {
	return ops.CompleteReview(ctx, c.b, id, reviewedBy)
}
