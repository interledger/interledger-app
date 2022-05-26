package fundingsources

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
	childLogger := logger.With(zap.String("service", "funding-sources"))
	return &loggingService{childLogger, service}
}

func (s *loggingService) Create(ctx context.Context, tx *sqlx.Tx, args *CreateArgs) (fs *FundingSource, err error) {
	defer func(begin time.Time) {
		if err != nil {
			s.logger.Error(
				"Failed to create funding source.",
				zap.Int64("took", time.Since(begin).Milliseconds()),
				zap.String("msg", err.Error()),
			)
			return
		}

		s.logger.Debug(
			"Created funding source.",
			zap.Int64("took", time.Since(begin).Milliseconds()),
		)
	}(time.Now())

	return s.Service.Create(ctx, tx, args)
}

func (s *loggingService) Get(ctx context.Context, id string) (fs *FundingSource, err error) {
	defer func(begin time.Time) {
		if err != nil {
			s.logger.Error(
				"Failed to get funding source.",
				zap.String("id", id),
				zap.Int64("took", time.Since(begin).Milliseconds()),
				zap.String("msg", err.Error()),
			)
			return
		}

		s.logger.Debug(
			"Got funding source.",
			zap.String("id", fs.ID),
			zap.Int64("took", time.Since(begin).Milliseconds()),
		)
	}(time.Now())
	return s.Service.Get(ctx, id)
}

func (s *loggingService) GetByAccountId(ctx context.Context, identityId string) (fs []FundingSource, err error) {
	defer func(begin time.Time) {
		if err != nil {
			s.logger.Error(
				"Failed to get funding sources.",
				zap.String("identityId", identityId),
				zap.Int64("took", time.Since(begin).Milliseconds()),
				zap.String("msg", err.Error()),
			)
			return
		}

		s.logger.Debug(
			"Got funding source.",
			// zap.String("id", fs[0]),
			zap.Int64("took", time.Since(begin).Milliseconds()),
		)
	}(time.Now())
	return s.Service.GetByAccountId(ctx, identityId)
}

func (s *loggingService) Verify(ctx context.Context, args *VerifyArgs) (fs *FundingSource, err error) {
	defer func(begin time.Time) {
		if err != nil {
			s.logger.Error(
				"Failed to verify funding source.",
				zap.String("id", args.FundingSourceID),
				zap.String("identityID", args.IdentityID),
				zap.Int64("took", time.Since(begin).Milliseconds()),
				zap.String("msg", err.Error()),
			)
			return
		}

		s.logger.Debug(
			"Verified funding source",
			zap.String("id", args.FundingSourceID),
			zap.String("identityID", args.IdentityID),
		)
	}(time.Now())

	return s.Service.Verify(ctx, args)
}

func (s *loggingService) CreateBankAccount(ctx context.Context, args *CreateBankAccountArgs) (fs *FundingSource, err error) {
	defer func(begin time.Time) {
		if err != nil {
			s.logger.Error(
				"Failed to link bank account.",
				zap.String("identityID", args.IdentityID),
				zap.String("accountID", args.AccountID),
				zap.Int64("took", time.Since(begin).Milliseconds()),
				zap.String("msg", err.Error()),
			)
			return
		}

		s.logger.Debug(
			"Linked Bank account",
			zap.String("id", fs.ID),
			zap.String("accountID", fs.AccountID),
			zap.String("identityID", args.IdentityID),
		)
	}(time.Now())

	return s.Service.CreateBankAccount(ctx, args)
}

func (s *loggingService) GetMxConnectWidget(ctx context.Context, accountID string, identityID string) (url string, err error) {
	defer func(begin time.Time) {
		if err != nil {
			s.logger.Error(
				"Failed to get mx connect widget.",
				zap.String("accountID", accountID),
				zap.Int64("took", time.Since(begin).Milliseconds()),
				zap.String("msg", err.Error()),
			)
			return
		}

		s.logger.Debug(
			"Got mx connect widget",
			zap.String("accountID", accountID),
		)
	}(time.Now())

	return s.Service.GetMxConnectWidget(ctx, accountID, identityID)
}

func (s *loggingService) CreateMxBankAccount(ctx context.Context, args *CreateMxBankAccountArgs) (fs *FundingSource, err error) {
	defer func(begin time.Time) {
		if err != nil {
			s.logger.Error(
				"Failed to create mx bank account.",
				zap.String("accountID", args.AccountID),
				zap.Int64("took", time.Since(begin).Milliseconds()),
				zap.String("msg", err.Error()),
			)
			return
		}

		s.logger.Debug(
			"Created mx bank account",
			zap.String("accountID", args.AccountID),
			zap.String("fundingSourceID", fs.ID),
		)
	}(time.Now())

	return s.Service.CreateMxBankAccount(ctx, args)
}
