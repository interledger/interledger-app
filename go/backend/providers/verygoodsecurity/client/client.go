package client

import (
	"context"
	"time"

	"gitlab.com/fynbos/backend/providers/verygoodsecurity"
	"gitlab.com/fynbos/backend/providers/verygoodsecurity/ops"
	"gitlab.com/fynbos/log"
	"go.uber.org/zap"
)

var _ verygoodsecurity.Client = client{}

type client struct {
	b ops.Backends
}

func New(b ops.Backends) verygoodsecurity.Client {
	return &client{
		b: b,
	}
}

func (c client) CreateCard(ctx context.Context, args verygoodsecurity.Card) (card *verygoodsecurity.Card, err error) {
	defer func(begin time.Time) {
		if err != nil {
			log.Error(
				"Failed to create vgs card.",
				zap.Int64("took", time.Since(begin).Milliseconds()),
				zap.String("msg", err.Error()),
			)
			return
		}

		log.Debug(
			"Created vgs card.",
			zap.Int64("took", time.Since(begin).Milliseconds()),
		)
	}(time.Now())

	return ops.CreateCard(ctx, c.b, args)
}
