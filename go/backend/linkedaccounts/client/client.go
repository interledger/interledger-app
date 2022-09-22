package client

import (
	"context"
	"time"

	"gitlab.com/fynbos/backend/linkedaccounts"
	"gitlab.com/fynbos/backend/linkedaccounts/ops"
	"go.uber.org/zap"
)

var _ linkedaccounts.Client = client{}

type client struct {
	logger *zap.Logger
	b      ops.Backends
}

func New(b ops.Backends, logger *zap.Logger) linkedaccounts.Client {
	return &client{
		b:      b,
		logger: logger.With(zap.String("ops", "linked-accounts")),
	}
}

func (c client) Create(ctx context.Context, args *linkedaccounts.CreateArgs) (fs *linkedaccounts.LinkedAccount, err error) {
	defer func(begin time.Time) {
		if err != nil {
			c.logger.Error(
				"Failed to create linked account.",
				zap.Int64("took", time.Since(begin).Milliseconds()),
				zap.String("msg", err.Error()),
			)
			return
		}

		c.logger.Debug(
			"Created linked account.",
			zap.Int64("took", time.Since(begin).Milliseconds()),
		)
	}(time.Now())

	return ops.Create(ctx, c.b, args)
}

func (c client) Get(ctx context.Context, id string) (fs *linkedaccounts.LinkedAccount, err error) {
	defer func(begin time.Time) {
		if err != nil {
			c.logger.Error(
				"Failed to get linked account.",
				zap.String("id", id),
				zap.Int64("took", time.Since(begin).Milliseconds()),
				zap.String("msg", err.Error()),
			)
			return
		}

		c.logger.Debug(
			"Got linked account.",
			zap.String("id", fs.ID),
			zap.Int64("took", time.Since(begin).Milliseconds()),
		)
	}(time.Now())

	return ops.Get(ctx, c.b, id)
}

func (c client) ListByWalletId(ctx context.Context, walletId string) (fsl []linkedaccounts.LinkedAccount, err error) {
	defer func(begin time.Time) {
		if err != nil {
			c.logger.Error(
				"Failed to get linked accounts.",
				zap.String("walletId", walletId),
				zap.Int64("took", time.Since(begin).Milliseconds()),
				zap.String("msg", err.Error()),
			)
			return
		}

		c.logger.Debug(
			"Got linked accounts.",
			// zap.String("id", fs[0]),
			zap.Int64("took", time.Since(begin).Milliseconds()),
		)
	}(time.Now())

	return ops.ListByWalletId(ctx, c.b, walletId)
}
