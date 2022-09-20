package client

import (
	"context"
	"time"

	"gitlab.com/fynbos/backend/fundingsources"
	"gitlab.com/fynbos/backend/fundingsources/ops"
	"go.uber.org/zap"
)

var _ fundingsources.Client = client{}

type client struct {
	logger *zap.Logger
	b      ops.Backends
}

func New(b ops.Backends, logger *zap.Logger) fundingsources.Client {
	return &client{
		b:      b,
		logger: logger.With(zap.String("ops", "funding-sources")),
	}
}

func (c client) Create(ctx context.Context, args *fundingsources.CreateArgs) (fs *fundingsources.FundingSource, err error) {
	defer func(begin time.Time) {
		if err != nil {
			c.logger.Error(
				"Failed to create funding source.",
				zap.Int64("took", time.Since(begin).Milliseconds()),
				zap.String("msg", err.Error()),
			)
			return
		}

		c.logger.Debug(
			"Created funding source.",
			zap.Int64("took", time.Since(begin).Milliseconds()),
		)
	}(time.Now())

	return ops.Create(ctx, c.b, args)
}

func (c client) Get(ctx context.Context, id string) (fs *fundingsources.FundingSource, err error) {
	defer func(begin time.Time) {
		if err != nil {
			c.logger.Error(
				"Failed to get funding source.",
				zap.String("id", id),
				zap.Int64("took", time.Since(begin).Milliseconds()),
				zap.String("msg", err.Error()),
			)
			return
		}

		c.logger.Debug(
			"Got funding source.",
			zap.String("id", fs.ID),
			zap.Int64("took", time.Since(begin).Milliseconds()),
		)
	}(time.Now())

	return ops.Get(ctx, c.b, id)
}

func (c client) GetByWalletId(ctx context.Context, walletId string) (fsl []fundingsources.FundingSource, err error) {
	defer func(begin time.Time) {
		if err != nil {
			c.logger.Error(
				"Failed to get funding sources.",
				zap.String("walletId", walletId),
				zap.Int64("took", time.Since(begin).Milliseconds()),
				zap.String("msg", err.Error()),
			)
			return
		}

		c.logger.Debug(
			"Got funding source.",
			// zap.String("id", fs[0]),
			zap.Int64("took", time.Since(begin).Milliseconds()),
		)
	}(time.Now())

	return ops.GetByWalletId(ctx, c.b, walletId)
}

func (c client) Verify(ctx context.Context, args *fundingsources.VerifyArgs) (fs *fundingsources.FundingSource, err error) {
	defer func(begin time.Time) {
		if err != nil {
			c.logger.Error(
				"Failed to verify funding source.",
				zap.String("id", args.FundingSourceID),
				zap.String("identityID", args.IdentityID),
				zap.Int64("took", time.Since(begin).Milliseconds()),
				zap.String("msg", err.Error()),
			)
			return
		}

		c.logger.Debug(
			"Verified funding source",
			zap.String("id", args.FundingSourceID),
			zap.String("identityID", args.IdentityID),
		)
	}(time.Now())

	return ops.Verify(ctx, c.b, args)
}
