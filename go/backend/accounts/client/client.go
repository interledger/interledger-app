package client

import (
	"context"
	"time"

	"gitlab.com/fynbos/backend/accounts/ops"

	"github.com/jmoiron/sqlx"
	"gitlab.com/fynbos/backend/accounts"
	"go.uber.org/zap"
)

var _ accounts.Client = client{}

type client struct {
	logger          *zap.Logger
	pacioliLedgerID uint32
	b               ops.Backends
}

func Make(b ops.Backends, logger *zap.Logger) accounts.Client {
	return &client{
		b:               b,
		pacioliLedgerID: 0,
		logger:          logger.With(zap.String("service", "account")),
	}
}
func (c client) Init(ctx context.Context) (err error) {

	/// Demo of what a logging interface could look like
	defer func(begin time.Time) {
		if err != nil {
			c.logger.Error(
				"Failed initialise.",
				zap.Int64("took", time.Since(begin).Milliseconds()),
				zap.String("msg", err.Error()),
			)
			return
		}

		c.logger.Debug(
			"Initialised.",
			zap.Int64("took", time.Since(begin).Milliseconds()),
		)
	}(time.Now())

	return ops.Init(ctx, c.b, c.pacioliLedgerID)
}

func (c client) Create(ctx context.Context, args accounts.CreateAccountArgs) (*accounts.Account, error) {
	return ops.Create(ctx, c.b, c.pacioliLedgerID, &args)

}

func (c client) GetByIdentityIDWithTrx(ctx context.Context, tx *sqlx.Tx, id string) (*accounts.Account, error) {
	return ops.GetByIdentityIDWithTrx(ctx, c.b, c.pacioliLedgerID, tx, id)
}

func (c client) GetByIdentityID(ctx context.Context, id string) (*accounts.Account, error) {
	return ops.GetByIdentityID(ctx, c.b, c.pacioliLedgerID, id)
}

func (c client) Get(ctx context.Context, id string) (*accounts.Account, error) {
	return ops.Get(ctx, c.b, c.pacioliLedgerID, id)
}

func (c client) VerifyWithTx(ctx context.Context, tx *sqlx.Tx, args accounts.VerifyArgs) (*accounts.Account, error) {
	return ops.VerifyWithTx(ctx, c.b, c.pacioliLedgerID, tx, &args)
}

func (c client) CanMakeOutgoingPayment(acc accounts.Account, identityID string) bool {
	return ops.CanMakeOutgoingPayment(&acc, identityID)
}

func (c client) CanMakeDeposit(acc accounts.Account, identityID string) bool {
	//TODO implement me
	panic("implement me")
}

func (c client) CanCreateFundingSource(acc accounts.Account, identityID string) bool {
	//TODO implement me
	panic("implement me")
}

func (c client) CanVerifyFundingSource(acc accounts.Account, identityID string) bool {
	//TODO implement me
	panic("implement me")
}
