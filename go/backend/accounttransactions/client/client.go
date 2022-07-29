package client

import (
	"context"
	"time"

	"github.com/jmoiron/sqlx"
	account_transactions "gitlab.com/fynbos/backend/accounttransactions"
	"gitlab.com/fynbos/backend/accounttransactions/ops"
	"go.uber.org/zap"
)

var _ account_transactions.Client = client{}

type client struct {
	logger *zap.Logger
	b      ops.Backends
}

func New(b ops.Backends, logger *zap.Logger) account_transactions.Client {
	return &client{
		b:      b,
		logger: logger.With(zap.String("ops", "account-transactions")),
	}

}

func (c client) Create(ctx context.Context, args *account_transactions.CreateTransactionArgs) (acc *account_transactions.AccountTransaction, err error) {
	defer func(begin time.Time) {
		if err != nil {
			c.logger.Error(
				"Failed to create account transaction.",
				zap.Int64("took", time.Since(begin).Milliseconds()),
				zap.String("msg", err.Error()),
			)
			return
		}

		c.logger.Debug(
			"Created account transaction.",
			zap.Int64("took", time.Since(begin).Milliseconds()),
			zap.String("type", args.Type),
			zap.Uint64("amount", args.NetAmount),
		)
	}(time.Now())

	return ops.Create(ctx, c.b, args)
}

func (c client) CreatePending(ctx context.Context, args *account_transactions.CreatePendingTransactionArgs) (at *account_transactions.AccountTransaction, err error) {
	defer func(begin time.Time) {
		if err != nil {
			c.logger.Error(
				"Failed to create account transaction.",
				zap.Int64("took", time.Since(begin).Milliseconds()),
				zap.String("msg", err.Error()),
			)
			return
		}

		c.logger.Debug(
			"Created account transaction.",
			zap.Int64("took", time.Since(begin).Milliseconds()),
			zap.String("type", args.Type),
			zap.Uint64("amount", args.NetAmount),
		)
	}(time.Now())

	return ops.CreatePending(ctx, c.b, args)
}

func (c client) PostPending(ctx context.Context, id string) (at *account_transactions.AccountTransaction, err error) {
	defer func(begin time.Time) {
		if err != nil {
			c.logger.Error(
				"Failed to post pending account transaction.",
				zap.Int64("took", time.Since(begin).Milliseconds()),
				zap.String("msg", err.Error()),
				zap.String("id", id),
			)
			return
		}

		c.logger.Debug(
			"posted pending account transaction.",
			zap.Int64("took", time.Since(begin).Milliseconds()),
			zap.String("id", id),
		)
	}(time.Now())

	return ops.PostPending(ctx, c.b, id)
}

func (c client) VoidPending(ctx context.Context, id string) (at *account_transactions.AccountTransaction, err error) {
	defer func(begin time.Time) {
		if err != nil {
			c.logger.Error(
				"Failed to void pending account transaction.",
				zap.Int64("took", time.Since(begin).Milliseconds()),
				zap.String("msg", err.Error()),
				zap.String("id", id),
			)
			return
		}

		c.logger.Debug(
			"voided pending account transaction.",
			zap.Int64("took", time.Since(begin).Milliseconds()),
			zap.String("id", id),
		)
	}(time.Now())

	return ops.VoidPending(ctx, c.b, id)
}

func (c client) GetByAccount(ctx context.Context, t *sqlx.Tx, args *account_transactions.GetByAccountArgs) (atl []*account_transactions.AccountTransaction, err error) {
	defer func(begin time.Time) {
		if err != nil {
			c.logger.Error(
				"Failed to get transactions for account.",
				zap.String("accountID", args.AccountID),
			)
			return
		}

		c.logger.Debug(
			"Got account transactions.",
			zap.String("accountID", args.AccountID),
			zap.Int64("took", time.Since(begin).Milliseconds()),
		)
	}(time.Now())

	return ops.GetByAccount(ctx, c.b, t, args)
}

func (c client) GetPage(ctx context.Context, args *account_transactions.PaginationArgs) (atl []account_transactions.AccountTransaction, err error) {
	defer func(begin time.Time) {
		if err != nil {
			c.logger.Error(
				"Failed to get account transaction page.",
				zap.String("accountID", args.AccountID),
				zap.String("after", args.After),
				zap.Uint32("first", args.First),
				zap.String("msg", err.Error()),
				zap.Int64("took", time.Since(begin).Milliseconds()),
			)
			return
		}

		c.logger.Debug(
			"Got account transaction page.",
			zap.String("accountID", args.AccountID),
			zap.String("after", args.After),
			zap.Uint32("first", args.First),
			zap.Int64("took", time.Since(begin).Milliseconds()),
		)
	}(time.Now())

	return ops.GetPage(ctx, c.b, args)
}

func (c client) GetPageInfo(ctx context.Context, accountID string, edges []account_transactions.AccountTransaction) (hasNextPage bool, startCursor string, endCursor string, err error) {
	defer func(begin time.Time) {
		if err != nil {
			c.logger.Error(
				"Failed to get account transaction page info.",
				zap.String("accountID", accountID),
				zap.String("msg", err.Error()),
				zap.Int64("took", time.Since(begin).Milliseconds()),
			)
			return
		}

		c.logger.Debug(
			"Got account transaction page info.",
			zap.String("accountID", accountID),
			zap.Int64("took", time.Since(begin).Milliseconds()),
		)
	}(time.Now())

	return ops.GetPageInfo(ctx, c.b, accountID, edges)
}

func (c client) Get(ctx context.Context, id string) (at *account_transactions.AccountTransaction, err error) {
	defer func(begin time.Time) {
		if err != nil {
			c.logger.Error(
				"Failed to get account transaction.",
				zap.String("id", id),
				zap.String("msg", err.Error()),
				zap.Int64("took", time.Since(begin).Milliseconds()),
			)
			return
		}

		c.logger.Debug(
			"Got account transaction.",
			zap.String("id", id),
			zap.Int64("took", time.Since(begin).Milliseconds()),
		)
	}(time.Now())

	return ops.Get(ctx, c.b, id)
}
