package account_transactions

import (
	"context"
	"time"

	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

type loggingService struct {
	logger  *zap.Logger
	Service Service
}

func NewLoggingService(service Service, logger *zap.Logger) Service {
	childLogger := logger.With(zap.String("service", "account-transactions"))
	return &loggingService{childLogger, service}
}

func (self *loggingService) Create(ctx context.Context, args *CreateTransactionArgs) (trx *AccountTransaction, err error) {
	defer func(begin time.Time) {
		if err != nil {
			self.logger.Error(
				"Failed to create account transaction.",
				zap.Int64("took", time.Since(begin).Milliseconds()),
				zap.String("msg", err.Error()),
			)
			return
		}

		self.logger.Debug(
			"Created account transaction.",
			zap.Int64("took", time.Since(begin).Milliseconds()),
			zap.String("type", args.Type),
			zap.Uint64("amount", args.NetAmount),
		)
	}(time.Now())

	return self.Service.Create(ctx, args)
}

func (self *loggingService) GetByAccount(ctx context.Context, tx *sqlx.Tx, args *GetByAccountArgs) (trxs []*AccountTransaction, err error) {
	defer func(begin time.Time) {
		if err != nil {
			self.logger.Error(
				"Failed to get transactions for account.",
				zap.String("accountID", args.AccountID),
			)
			return
		}

		self.logger.Debug(
			"Got account transactions.",
			zap.String("accountID", args.AccountID),
			zap.Int64("took", time.Since(begin).Milliseconds()),
		)
	}(time.Now())

	return self.Service.GetByAccount(ctx, tx, args)
}

func (self *loggingService) GetPage(
	ctx context.Context,
	args *PaginationArgs,
) (edges []AccountTransaction, err error) {
	defer func(begin time.Time) {
		if err != nil {
			self.logger.Error(
				"Failed to get account transaction page.",
				zap.String("accountID", args.AccountID),
				zap.String("after", args.After),
				zap.Uint32("first", args.First),
				zap.String("msg", err.Error()),
				zap.Int64("took", time.Since(begin).Milliseconds()),
			)
			return
		}

		self.logger.Debug(
			"Got account transaction page.",
			zap.String("accountID", args.AccountID),
			zap.String("after", args.After),
			zap.Uint32("first", args.First),
			zap.Int64("took", time.Since(begin).Milliseconds()),
		)
	}(time.Now())

	return self.Service.GetPage(ctx, args)
}

func (self *loggingService) GetPageInfo(
	ctx context.Context,
	accountID string,
	edges []AccountTransaction,
) (hasNextPage bool, startCursor string, endCursor string, err error) {
	defer func(begin time.Time) {
		if err != nil {
			self.logger.Error(
				"Failed to get account transaction page info.",
				zap.String("accountID", accountID),
				zap.String("msg", err.Error()),
				zap.Int64("took", time.Since(begin).Milliseconds()),
			)
			return
		}

		self.logger.Debug(
			"Got account transaction page info.",
			zap.String("accountID", accountID),
			zap.Int64("took", time.Since(begin).Milliseconds()),
		)
	}(time.Now())

	return self.Service.GetPageInfo(ctx, accountID, edges)
}

func (self *loggingService) Get(ctx context.Context, id string) (trx *AccountTransaction, err error) {
	defer func(begin time.Time) {
		if err != nil {
			self.logger.Error(
				"Failed to get account transaction.",
				zap.String("id", id),
				zap.String("msg", err.Error()),
				zap.Int64("took", time.Since(begin).Milliseconds()),
			)
			return
		}

		self.logger.Debug(
			"Got account transaction.",
			zap.String("id", id),
			zap.Int64("took", time.Since(begin).Milliseconds()),
		)
	}(time.Now())

	return self.Service.Get(ctx, id)
}
