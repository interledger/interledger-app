package client

import (
	"context"
	"time"

	"github.com/jmoiron/sqlx"
	"gitlab.com/fynbos/backend/accounts"
	"gitlab.com/fynbos/backend/accounts/ops"
	"go.uber.org/zap"
)

var _ accounts.Client = client{}

type client struct {
	logger          *zap.Logger
	pacioliLedgerID uint32
	b               ops.Backends
}

func New(b ops.Backends, ledgerID uint32, logger *zap.Logger) accounts.Client {
	return &client{
		b:               b,
		pacioliLedgerID: ledgerID,
		logger:          logger.With(zap.String("ops", "account")),
	}
}

func (c client) Create(ctx context.Context, args *accounts.CreateAccountArgs) (acc *accounts.Account, err error) {
	defer func(begin time.Time) {
		if err != nil {
			c.logger.Error(
				"Failed to create account.",
				zap.Int64("took", time.Since(begin).Milliseconds()),
				zap.String("msg", err.Error()),
			)
			return
		}

		c.logger.Debug(
			"Created account.",
			zap.Int64("took", time.Since(begin).Milliseconds()),
		)
	}(time.Now())

	c.logger.Debug(
		"Creating account.",
		zap.String("identityID", args.IdentityID),
	)
	return ops.Create(ctx, c.b, c.pacioliLedgerID, args)
}

func (c client) GetByIdentityIDWithTrx(ctx context.Context, tx *sqlx.Tx, id string) (acc *accounts.Account, err error) {
	defer func(begin time.Time) {
		if err != nil {
			c.logger.Error(
				"Failed to get account.",
				zap.String("identityID", id),
				zap.Int64("took", time.Since(begin).Milliseconds()),
				zap.String("msg", err.Error()),
			)
			return
		}

		c.logger.Debug(
			"Got account.",
			zap.String("id", acc.ID),
			zap.Int64("took", time.Since(begin).Milliseconds()),
		)
	}(time.Now())

	return ops.GetByIdentityIDWithTrx(ctx, c.b, tx, id)
}

func (c client) GetByIdentityID(ctx context.Context, id string) (acc *accounts.Account, err error) {
	defer func(begin time.Time) {
		if err != nil {
			c.logger.Error(
				"Failed to get account.",
				zap.String("identityID", id),
				zap.Int64("took", time.Since(begin).Milliseconds()),
				zap.String("msg", err.Error()),
			)
			return
		}

		c.logger.Debug(
			"Got account.",
			zap.String("id", acc.ID),
			zap.Int64("took", time.Since(begin).Milliseconds()),
		)
	}(time.Now())

	return ops.GetByIdentityID(ctx, c.b, id)
}

func (c client) Get(ctx context.Context, id string) (acc *accounts.Account, err error) {
	defer func(begin time.Time) {
		if err != nil {
			c.logger.Error(
				"Failed to get account.",
				zap.String("accountID", id),
				zap.Int64("took", time.Since(begin).Milliseconds()),
				zap.String("msg", err.Error()),
			)
			return
		}

		c.logger.Debug(
			"Got account.",
			zap.String("id", acc.ID),
			zap.Int64("took", time.Since(begin).Milliseconds()),
		)
	}(time.Now())

	return ops.Get(ctx, c.b, id)
}

func (c client) CanMakeOutgoingPayment(acc *accounts.Account, identityID string) bool {
	return ops.CanMakeOutgoingPayment(acc, identityID)
}

func (c client) CanMakeDeposit(acc *accounts.Account, identityID string) bool {
	return ops.CanMakeDeposit(acc, identityID)
}

func (c client) CanCreateFundingSource(acc *accounts.Account, identityID string) bool {
	return ops.CanCreateFundingSource(acc, identityID)
}

func (c client) CanVerifyFundingSource(acc *accounts.Account, identityID string) bool {
	return ops.CanVerifyFundingSource(acc, identityID)
}
