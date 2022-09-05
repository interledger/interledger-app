package client

import (
	"context"
	"time"

	"gitlab.com/fynbos/backend/payments"
	"gitlab.com/fynbos/backend/payments/ops"
	"go.uber.org/zap"
)

var _ payments.Client = client{}

type client struct {
	logger *zap.Logger
	b      ops.Backends
}

func New(b ops.Backends, logger *zap.Logger) payments.Client {
	return &client{
		b:      b,
		logger: logger.With(zap.String("ops", "payments")),
	}
}

func (c client) InitiateOutgoingPayment(ctx context.Context, args payments.InitiateOutgoingPaymentArgs) (outgoingPayment *payments.OutgoingPayment, err error) {
	defer func(begin time.Time) {
		if err != nil {
			c.logger.Error(
				"Failed to initiate outgoing payment.",
				zap.String("userID", args.UserID),
				zap.String("to", args.To),
				zap.Uint64("amount", args.Amount),
				zap.String("msg", err.Error()),
			)
			return
		}

		c.logger.Debug(
			"Initiated outgoing payment.",
			zap.String("userID", args.UserID),
			zap.String("to", args.To),
			zap.Uint64("amount", args.Amount),
			zap.Int64("took", time.Since(begin).Milliseconds()),
		)
	}(time.Now())

	return ops.InitiateOutgoingPayment(ctx, c.b, args)
}

func (c client) Get(ctx context.Context, id string) (outgoingPayment *payments.OutgoingPayment, err error) {
	defer func(begin time.Time) {
		if err != nil {
			c.logger.Error(
				"Failed to get outgoing payment.",
				zap.String("id", id),
				zap.String("msg", err.Error()),
			)
			return
		}

		c.logger.Debug(
			"Got outgoing payment.",
			zap.String("id", id),
			zap.Int64("took", time.Since(begin).Milliseconds()),
		)
	}(time.Now())

	return ops.Get(ctx, c.b, id)
}

func (c client) SetState(ctx context.Context, id string, state payments.State) error {
	defer func(begin time.Time) {
		c.logger.Debug(
			"Set outgoing payment state.",
			zap.String("id", id),
			zap.String("state", state.String()),
			zap.Int64("took", time.Since(begin).Milliseconds()),
		)
	}(time.Now())

	return ops.SetState(ctx, c.b, id, state)
}
