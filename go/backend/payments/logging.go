package payments

import (
	"context"
	"time"

	account_transactions "gitlab.com/fynbos/backend/accounttransactions"
	"go.uber.org/zap"
)

type loggingService struct {
	logger  *zap.Logger
	Service Service
}

func NewLoggingService(svc Service, logger *zap.Logger) Service {
	childLogger := logger.With(zap.String("service", "account"))
	return &loggingService{childLogger, svc}
}

func (s *loggingService) InitiateOutgoingPayment(
	ctx context.Context,
	args *InitiateOutgoingPaymentArgs,
) (acc *account_transactions.AccountTransaction, err error) {
	defer func(begin time.Time) {
		if err != nil {
			s.logger.Error(
				"Failed to initiate outgoing payment.",
				zap.String("identityID", args.IdentityID),
				zap.String("accountID", args.AccountID),
				zap.String("to", args.To),
				zap.Uint64("amount", args.Amount),
				zap.String("msg", err.Error()),
			)
			return
		}

		s.logger.Debug(
			"Initiated outgoing payment.",
			zap.String("identityID", args.IdentityID),
			zap.String("accountID", args.AccountID),
			zap.String("to", args.To),
			zap.Uint64("amount", args.Amount),
			zap.Int64("took", time.Since(begin).Milliseconds()),
		)
	}(time.Now())

	return s.Service.InitiateOutgoingPayment(ctx, args)
}
