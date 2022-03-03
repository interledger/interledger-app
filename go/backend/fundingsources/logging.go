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

func (self *loggingService) Create(ctx context.Context, tx *sqlx.Tx, args *CreateArgs) (fs *FundingSource, err error) {
	defer func(begin time.Time) {
		if err != nil {
			self.logger.Error(
				"Failed to create funding source.",
				zap.Int64("took", time.Since(begin).Milliseconds()),
				zap.String("msg", err.Error()),
			)
			return
		}

		self.logger.Debug(
			"Created funding source.",
			zap.Int64("took", time.Since(begin).Milliseconds()),
		)
	}(time.Now())

	return self.Service.Create(ctx, tx, args)
}

func (self *loggingService) Get(ctx context.Context, tx *sqlx.Tx, id string) (fs *FundingSource, err error) {
	defer func(begin time.Time) {
		if err != nil {
			self.logger.Error(
				"Failed to get funding source.",
				zap.String("id", id),
				zap.Int64("took", time.Since(begin).Milliseconds()),
				zap.String("msg", err.Error()),
			)
			return
		}

		self.logger.Debug(
			"Got funding source.",
			zap.String("id", fs.ID),
			zap.Int64("took", time.Since(begin).Milliseconds()),
		)
	}(time.Now())
	return self.Service.Get(ctx, tx, id)
}

func (self *loggingService) GetByAccountId(ctx context.Context, tx *sqlx.Tx, identityId string) (fs []FundingSource, err error) {
	defer func(begin time.Time) {
		if err != nil {
			self.logger.Error(
				"Failed to get funding sources.",
				zap.String("identityId", identityId),
				zap.Int64("took", time.Since(begin).Milliseconds()),
				zap.String("msg", err.Error()),
			)
			return
		}

		self.logger.Debug(
			"Got funding source.",
			// zap.String("id", fs[0]),
			zap.Int64("took", time.Since(begin).Milliseconds()),
		)
	}(time.Now())
	return self.Service.GetByAccountId(ctx, tx, identityId)
}

func (self *loggingService) Verify(ctx context.Context, args *VerifyArgs) (fs *FundingSource, err error) {
	defer func(begin time.Time) {
		if err != nil {
			self.logger.Error(
				"Failed to verify funding source.",
				zap.String("id", args.FundingSourceID),
				zap.String("identityID", args.IdentityID),
				zap.Int64("took", time.Since(begin).Milliseconds()),
				zap.String("msg", err.Error()),
			)
			return
		}

		self.logger.Debug(
			"Verified funding source",
			zap.String("id", args.FundingSourceID),
			zap.String("identityID", args.IdentityID),
		)
	}(time.Now())

	return self.Service.Verify(ctx, args)
}

func (self *loggingService) CreateBankAccount(ctx context.Context, args *CreateBankAccountArgs) (fs *FundingSource, err error) {
	defer func(begin time.Time) {
		if err != nil {
			self.logger.Error(
				"Failed to link bank account.",
				zap.String("identityID", args.IdentityID),
				zap.String("accountID", args.AccountID),
				zap.Int64("took", time.Since(begin).Milliseconds()),
				zap.String("msg", err.Error()),
			)
			return
		}

		self.logger.Debug(
			"Linked Bank account",
			zap.String("id", fs.ID),
			zap.String("accountID", fs.AccountID),
			zap.String("identityID", args.IdentityID),
		)
	}(time.Now())

	return self.Service.CreateBankAccount(ctx, args)
}
