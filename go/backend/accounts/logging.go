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

func (self *loggingService) Init(ctx context.Context) (err error) {
	defer func(begin time.Time) {
		if err != nil {
			self.logger.Error(
				"Failed initialise.",
				zap.Int64("took", time.Since(begin).Milliseconds()),
				zap.String("msg", err.Error()),
			)
			return
		}

		self.logger.Debug(
			"Initialised.",
			zap.Int64("took", time.Since(begin).Milliseconds()),
		)
	}(time.Now())

	return self.Service.Init(ctx)
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

func (s *loggingService) VerifyWithTx(ctx context.Context, tx *sqlx.Tx, args *VerifyArgs) (account *Account, err error) {
	defer func(begin time.Time) {
		if err != nil {
			s.logger.Error(
				"Failed to verify account.",
				zap.String("accountID", args.AccountID),
				zap.Int64("took", time.Since(begin).Milliseconds()),
				zap.String("msg", err.Error()),
			)
			return
		}

		s.logger.Debug(
			"Account verified.",
			zap.String("id", account.ID),
			zap.Int64("took", time.Since(begin).Milliseconds()),
		)
	}(time.Now())

	return s.Service.VerifyWithTx(ctx, tx, args)
}

func (s loggingService) CanMakeOutgoingPayment(acc *Account, identityID string) bool {
	return s.Service.CanMakeOutgoingPayment(acc, identityID)
}

func (s loggingService) CanMakeDeposit(acc *Account, identityID string) bool {
	return s.Service.CanMakeDeposit(acc, identityID)
}

func (s loggingService) CanCreateFundingSource(acc *Account, identityID string) bool {
	return s.Service.CanCreateFundingSource(acc, identityID)
}

func (s loggingService) CanVerifyFundingSource(acc *Account, identityID string) bool {
	return s.Service.CanVerifyFundingSource(acc, identityID)
}
