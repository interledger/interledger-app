package client

import (
	"context"
	"time"

	"gitlab.com/fynbos/backend/waitlist"
	"gitlab.com/fynbos/backend/waitlist/ops"
	"go.uber.org/zap"
)

var _ waitlist.Client = client{}

type client struct {
	b      ops.Backends
	logger *zap.Logger
}

func New(b ops.Backends, logger zap.Logger) waitlist.Client {
	return &client{
		b:      b,
		logger: logger.With(zap.String("service", "waitlist")),
	}
}

func (c client) Add(ctx context.Context, email, countryCode string) (err error) {
	defer func(begin time.Time) {
		if err != nil {
			c.logger.Error(
				"Failed to add user to waitlist",
				zap.Int64("took", time.Since(begin).Milliseconds()),
				zap.String("msg", err.Error()),
			)
			return
		}

		c.logger.Debug(
			"email address added to waitlist",
			zap.Int64("took", time.Since(begin).Milliseconds()),
		)
	}(time.Now())

	return ops.AddSignup(ctx, c.b, email, countryCode)
}
