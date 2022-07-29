package client

import (
	"context"
	"time"

	"gitlab.com/fynbos/backend/identity"
	"gitlab.com/fynbos/backend/identity/ops"
	"go.uber.org/zap"
)

var _ identity.Client = client{}

type client struct {
	logger *zap.Logger
	b      ops.Backends
}

func New(b ops.Backends, logger *zap.Logger) identity.Client {
	return &client{
		b:      b,
		logger: logger.With(zap.String("ops", "identity")),
	}
}

func (c client) Create(ctx context.Context, args *identity.CreateArgs) (id *identity.Identity, err error) {
	defer func(begin time.Time) {
		if err != nil {
			c.logger.Error(
				"Failed to create identity.",
				zap.Int64("took", time.Since(begin).Milliseconds()),
				zap.String("msg", err.Error()),
			)
			return
		}

		c.logger.Debug(
			"Created identity.",
			zap.Int64("took", time.Since(begin).Milliseconds()),
		)
	}(time.Now())

	return ops.Create(ctx, c.b, args)
}

func (c client) Get(ctx context.Context, identityID string) (id *identity.Identity, err error) {
	defer func(begin time.Time) {
		if err != nil {
			c.logger.Error(
				"Failed to get identity.",
				zap.String("id", identityID),
				zap.Int64("took", time.Since(begin).Milliseconds()),
				zap.String("msg", err.Error()),
			)
			return
		}

		c.logger.Debug(
			"Got identity.",
			zap.String("id", id.ID),
			zap.Int64("took", time.Since(begin).Milliseconds()),
		)
	}(time.Now())

	return ops.Get(ctx, c.b, identityID)
}

func (c client) GetByEmail(ctx context.Context, email string) (id *identity.Identity, err error) {
	defer func(begin time.Time) {
		if err != nil {
			c.logger.Error(
				"Failed to get identity by email.",
				zap.String("email", email),
				zap.Int64("took", time.Since(begin).Milliseconds()),
				zap.String("msg", err.Error()),
			)
			return
		}

		c.logger.Debug(
			"Got identity.",
			zap.String("id", id.ID),
			zap.Int64("took", time.Since(begin).Milliseconds()),
		)
	}(time.Now())

	return ops.GetByEmail(ctx, c.b, email)
}
