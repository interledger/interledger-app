package accounts

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

func NewLoggingService(accountsService Service, logger *zap.Logger) Service {
	childLogger := logger.With(zap.String("service", "account"))
	return &loggingService{childLogger, accountsService}
}

func (self *loggingService) Create(ctx context.Context, tx *sqlx.Tx, args *CreateAccountArgs) (account *Account, err error) {
	defer func(begin time.Time) {
		if err != nil {
			self.logger.Error(
				"Failed to create account.",
				zap.Int64("took", time.Since(begin).Milliseconds()),
				zap.String("msg", err.Error()),
			)
			return
		}

		self.logger.Debug(
			"Created account.",
			zap.Int64("took", time.Since(begin).Milliseconds()),
		)
	}(time.Now())

	self.logger.Debug(
		"Creating account.",
		zap.String("identityID", args.IdentityID),
		zap.String("country", args.Country),
	)
	return self.Service.Create(ctx, tx, args)
}

func (self *loggingService) GetByIdentityIDWithTrx(ctx context.Context, tx *sqlx.Tx, identityID string) (account *Account, err error) {
	defer func(begin time.Time) {
		if err != nil {
			self.logger.Error(
				"Failed to get account.",
				zap.String("identityID", identityID),
				zap.Int64("took", time.Since(begin).Milliseconds()),
				zap.String("msg", err.Error()),
			)
			return
		}

		self.logger.Debug(
			"Got account.",
			zap.String("id", account.ID),
			zap.Int64("took", time.Since(begin).Milliseconds()),
		)
	}(time.Now())

	return self.Service.GetByIdentityIDWithTrx(ctx, tx, identityID)
}

func (self *loggingService) GetByIdentityID(ctx context.Context, identityID string) (account *Account, err error) {
	defer func(begin time.Time) {
		if err != nil {
			self.logger.Error(
				"Failed to get account.",
				zap.String("identityID", identityID),
				zap.Int64("took", time.Since(begin).Milliseconds()),
				zap.String("msg", err.Error()),
			)
			return
		}

		self.logger.Debug(
			"Got account.",
			zap.String("id", account.ID),
			zap.Int64("took", time.Since(begin).Milliseconds()),
		)
	}(time.Now())

	return self.Service.GetByIdentityID(ctx, identityID)
}

func (self *loggingService) Get(ctx context.Context, tx *sqlx.Tx, accountID string) (account *Account, err error) {
	defer func(begin time.Time) {
		if err != nil {
			self.logger.Error(
				"Failed to get account.",
				zap.String("accountID", accountID),
				zap.Int64("took", time.Since(begin).Milliseconds()),
				zap.String("msg", err.Error()),
			)
			return
		}

		self.logger.Debug(
			"Got account.",
			zap.String("id", account.ID),
			zap.Int64("took", time.Since(begin).Milliseconds()),
		)
	}(time.Now())

	return self.Service.Get(ctx, tx, accountID)
}
