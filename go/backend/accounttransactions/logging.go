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

func (s *loggingService) Create(ctx context.Context, args *CreateTransactionArgs) (trx *AccountTransaction, err error) {
	defer func(begin time.Time) {
		if err != nil {
			s.logger.Error(
				"Failed to create account transaction.",
				zap.Int64("took", time.Since(begin).Milliseconds()),
				zap.String("msg", err.Error()),
			)
			return
		}

		s.logger.Debug(
			"Created account transaction.",
			zap.Int64("took", time.Since(begin).Milliseconds()),
			zap.String("type", args.Type),
			zap.Uint64("amount", args.NetAmount),
		)
	}(time.Now())

	return s.Service.Create(ctx, args)
}

func (s *loggingService) CreatePending(ctx context.Context, args *CreatePendingTransactionArgs) (trx *AccountTransaction, err error) {
	defer func(begin time.Time) {
		if err != nil {
			s.logger.Error(
				"Failed to create account transaction.",
				zap.Int64("took", time.Since(begin).Milliseconds()),
				zap.String("msg", err.Error()),
			)
			return
		}

		s.logger.Debug(
			"Created account transaction.",
			zap.Int64("took", time.Since(begin).Milliseconds()),
			zap.String("type", args.Type),
			zap.Uint64("amount", args.NetAmount),
		)
	}(time.Now())

	return s.Service.CreatePending(ctx, args)
}

func (s *loggingService) PostPending(ctx context.Context, id string) (trx *AccountTransaction, err error) {
	defer func(begin time.Time) {
		if err != nil {
			s.logger.Error(
				"Failed to post pending account transaction.",
				zap.Int64("took", time.Since(begin).Milliseconds()),
				zap.String("msg", err.Error()),
				zap.String("id", id),
			)
			return
		}

		s.logger.Debug(
			"posted pending account transaction.",
			zap.Int64("took", time.Since(begin).Milliseconds()),
			zap.String("id", id),
		)
	}(time.Now())

	return s.Service.PostPending(ctx, id)
}

func (s *loggingService) VoidPending(ctx context.Context, id string) (trx *AccountTransaction, err error) {
	defer func(begin time.Time) {
		if err != nil {
			s.logger.Error(
				"Failed to void pending account transaction.",
				zap.Int64("took", time.Since(begin).Milliseconds()),
				zap.String("msg", err.Error()),
				zap.String("id", id),
			)
			return
		}

		s.logger.Debug(
			"voided pending account transaction.",
			zap.Int64("took", time.Since(begin).Milliseconds()),
			zap.String("id", id),
		)
	}(time.Now())

	return s.Service.VoidPending(ctx, id)
}

func (s *loggingService) GetByAccount(ctx context.Context, tx *sqlx.Tx, args *GetByAccountArgs) (trxs []*AccountTransaction, err error) {
	defer func(begin time.Time) {
		if err != nil {
			s.logger.Error(
				"Failed to get transactions for account.",
				zap.String("accountID", args.AccountID),
			)
			return
		}

		s.logger.Debug(
			"Got account transactions.",
			zap.String("accountID", args.AccountID),
			zap.Int64("took", time.Since(begin).Milliseconds()),
		)
	}(time.Now())

	return s.Service.GetByAccount(ctx, tx, args)
}

func (s *loggingService) GetPage(
	ctx context.Context,
	args *PaginationArgs,
) (edges []AccountTransaction, err error) {
	defer func(begin time.Time) {
		if err != nil {
			s.logger.Error(
				"Failed to get account transaction page.",
				zap.String("accountID", args.AccountID),
				zap.String("after", args.After),
				zap.Uint32("first", args.First),
				zap.String("msg", err.Error()),
				zap.Int64("took", time.Since(begin).Milliseconds()),
			)
			return
		}

		s.logger.Debug(
			"Got account transaction page.",
			zap.String("accountID", args.AccountID),
			zap.String("after", args.After),
			zap.Uint32("first", args.First),
			zap.Int64("took", time.Since(begin).Milliseconds()),
		)
	}(time.Now())

	return s.Service.GetPage(ctx, args)
}

func (s *loggingService) GetPageInfo(
	ctx context.Context,
	accountID string,
	edges []AccountTransaction,
) (hasNextPage bool, startCursor string, endCursor string, err error) {
	defer func(begin time.Time) {
		if err != nil {
			s.logger.Error(
				"Failed to get account transaction page info.",
				zap.String("accountID", accountID),
				zap.String("msg", err.Error()),
				zap.Int64("took", time.Since(begin).Milliseconds()),
			)
			return
		}

		s.logger.Debug(
			"Got account transaction page info.",
			zap.String("accountID", accountID),
			zap.Int64("took", time.Since(begin).Milliseconds()),
		)
	}(time.Now())

	return s.Service.GetPageInfo(ctx, accountID, edges)
}

func (s *loggingService) Get(ctx context.Context, id string) (trx *AccountTransaction, err error) {
	defer func(begin time.Time) {
		if err != nil {
			s.logger.Error(
				"Failed to get account transaction.",
				zap.String("id", id),
				zap.String("msg", err.Error()),
				zap.Int64("took", time.Since(begin).Milliseconds()),
			)
			return
		}

		s.logger.Debug(
			"Got account transaction.",
			zap.String("id", id),
			zap.Int64("took", time.Since(begin).Milliseconds()),
		)
	}(time.Now())

	return s.Service.Get(ctx, id)
}
