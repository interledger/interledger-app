package fundingsources

import (
	"context"
	"time"

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

func (s *loggingService) Create(ctx context.Context, args *CreateArgs) (fs *FundingSource, err error) {
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

	return s.Service.Create(ctx, args)
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

func (s *loggingService) VerifyMxBankAccount(
	ctx context.Context,
	identityID string,
	fundingsourceID string,
) (fs *FundingSource, err error) {
	defer func(begin time.Time) {
		if err != nil {
			s.logger.Error(
				"Failed to verify mx bank account.",
				zap.String("identityID", identityID),
				zap.String("fundingSourceID", fs.ID),
				zap.Int64("took", time.Since(begin).Milliseconds()),
				zap.String("msg", err.Error()),
			)
			return
		}

		s.logger.Debug(
			"Verified mx bank account",
			zap.String("identityID", identityID),
			zap.String("fundingSourceID", fs.ID),
		)
	}(time.Now())

	return s.Service.VerifyMxBankAccount(ctx, identityID, fundingsourceID)
}

func (s *loggingService) CreateUnitCounterPartyFromMxAccount(
	ctx context.Context,
	fundingsourceID string,
) (cp *UnitCounterParty, err error) {
	defer func() {
		if err != nil {
			s.logger.Error(
				"Failed to create unit counterparty from mx bank account.",
				zap.String("fundingsourceID", fundingsourceID),
				zap.String("msg", err.Error()),
			)
			return
		}

		s.logger.Debug(
			"Created unit counterparty from mx account.",
			zap.String("fundingsourceID", fundingsourceID),
		)
	}()

	return s.Service.CreateUnitCounterPartyFromMxAccount(ctx, fundingsourceID)
}

func (s *loggingService) SetMxFundingSourceMask(ctx context.Context, fundingsourceID string) (err error) {
	defer func() {
		if err != nil {
			s.logger.Error(
				"Failed to set mx fundingsource mask.",
				zap.String("fundingsourceID", fundingsourceID),
				zap.String("msg", err.Error()),
			)
			return
		}

		s.logger.Debug(
			"Set mx fundingsource mask.",
			zap.String("fundingsourceID", fundingsourceID),
		)
	}()

	return s.Service.SetMxFundingSourceMask(ctx, fundingsourceID)
}

func (s *loggingService) SetMask(ctx context.Context, fundingsourceID string, mask string) (fs *FundingSource, err error) {
	defer func() {
		if err != nil {
			s.logger.Error(
				"Failed to set mask.",
				zap.String("fundingsourceID", fundingsourceID),
				zap.String("msg", err.Error()),
			)
			return
		}

		s.logger.Debug(
			"Set mask.",
			zap.String("fundingsourceID", fundingsourceID),
		)
	}()

	return s.Service.SetMask(ctx, fundingsourceID, mask)
}
