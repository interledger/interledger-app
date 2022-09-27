package client

import (
	"context"

	"gitlab.com/fynbos/backend/providers/fakecash"
	"gitlab.com/fynbos/backend/providers/fakecash/ops"
	"gitlab.com/fynbos/env"
	"gitlab.com/fynbos/pacioli"
	"go.uber.org/zap"
)

var _ fakecash.Client = client{}

type client struct {
	logger *zap.Logger
	b      ops.Backends
}

func New(b Backends, logger *zap.Logger) fakecash.Client {
	if env.IsProd() {
		panic("Fake cash not for prod.")
	}

	opsBackends := opsBackends{
		ledgerID: 4294967295,
		Backends: b,
	}

	logger.Info("Configuring fakecash-usd ledger...")
	ledgerErrors, err := b.Pacioli().ConfigureLedgers(context.Background(), []pacioli.ConfigureLedgerArgs{
		{
			ID:    opsBackends.LedgerID(),
			Name:  "fakecash-usd",
			Asset: "USD",
			Scale: 2,
		},
	})
	if err != nil {
		logger.Error("Failed to configure fakecash-usd ledger", zap.Error(err))
	}
	if len(ledgerErrors) > 1 {
		logger.Error("Failed to configure fakecash-usd ledger", zap.String("error", ledgerErrors[0].Code.String()))
	}

	return &client{
		b:      opsBackends,
		logger: logger.With(zap.String("ops", "fake-cash")),
	}
}

func (c client) Create(ctx context.Context, args fakecash.CreateArgs) (*fakecash.Account, error) {
	return ops.Create(ctx, c.b, args)
}

func (c client) Get(ctx context.Context, id string) (*fakecash.Account, error) {
	return ops.Get(ctx, c.b, id)
}
