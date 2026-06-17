package client

import (
	"context"
	"time"

	"github.com/interledger/interledger-app/go/backend/waitlist"
	"github.com/interledger/interledger-app/go/backend/waitlist/ops"
	"go.uber.org/zap"
)

var _ waitlist.Client = client{}

type client struct {
	b      ops.Backends
	logger *zap.Logger
}

func New(b ops.Backends, logger *zap.Logger) waitlist.Client {
	return &client{
		b:      b,
		logger: logger.With(zap.String("service", "waitlist")),
	}
}

func (c client) Add(ctx context.Context, email, countryCode, fullName, mugID string, betaOptIn bool) (err error) {
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

	return ops.AddSignup(ctx, c.b, email, countryCode, fullName, mugID, betaOptIn)
}

func (c client) IsMugAvailable(ctx context.Context, mugID string) (bool, error) {
	return ops.IsMugAvailable(ctx, c.b, mugID)
}

func (c client) CanSignup(ctx context.Context, id string) (bool bool, err error) {
	defer func(begin time.Time) {
		if err != nil {
			c.logger.Error(
				"Failed to check if id can signup",
				zap.Int64("took", time.Since(begin).Milliseconds()),
				zap.String("msg", err.Error()),
			)
			return
		}

		c.logger.Debug(
			" address added to waitlist",
			zap.Int64("took", time.Since(begin).Milliseconds()),
		)
	}(time.Now())

	return ops.CanSignup(ctx, c.b, id)
}

func (c client) SetSignupComplete(ctx context.Context, id, userID string) (err error) {
	defer func(begin time.Time) {
		if err != nil {
			c.logger.Error(
				"Failed to set signup complete",
				zap.Int64("took", time.Since(begin).Milliseconds()),
				zap.String("msg", err.Error()),
			)
			return
		}

		c.logger.Debug(
			"set signup as complete",
			zap.Int64("took", time.Since(begin).Milliseconds()),
		)
	}(time.Now())

	return ops.SetSignupComplete(ctx, c.b, id, userID)
}

func (c client) ListSignups(ctx context.Context) ([]waitlist.Signup, error) {
	return ops.ListSignups(ctx, c.b)
}

func (c client) AllowSignupById(ctx context.Context, id string) error {
	return ops.AllowSignupById(ctx, c.b, id)
}
