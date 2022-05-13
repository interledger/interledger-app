package identity

import (
	"context"
	"time"

	"go.uber.org/zap"
)

type loggingService struct {
	logger  *zap.Logger
	Service Service
}

func NewLoggingService(identityService Service, logger *zap.Logger) Service {
	childLogger := logger.With(zap.String("service", "identity"))
	return &loggingService{childLogger, identityService}
}

func (s *loggingService) Create(ctx context.Context, args *CreateArgs) (identity *Identity, err error) {
	defer func(begin time.Time) {
		if err != nil {
			s.logger.Error(
				"Failed to create identity.",
				zap.Int64("took", time.Since(begin).Milliseconds()),
				zap.String("msg", err.Error()),
			)
			return
		}

		s.logger.Debug(
			"Created identity.",
			zap.Int64("took", time.Since(begin).Milliseconds()),
		)
	}(time.Now())

	s.logger.Debug(
		"Creating identity.",
		zap.Stringer("args", args),
	)
	return s.Service.Create(ctx, args)
}

func (s *loggingService) Get(ctx context.Context, id string) (identity *Identity, err error) {
	defer func(begin time.Time) {
		if err != nil {
			s.logger.Error(
				"Failed to get identity.",
				zap.String("id", id),
				zap.Int64("took", time.Since(begin).Milliseconds()),
				zap.String("msg", err.Error()),
			)
			return
		}

		s.logger.Debug(
			"Got identity.",
			zap.String("id", identity.ID),
			zap.Int64("took", time.Since(begin).Milliseconds()),
		)
	}(time.Now())

	return s.Service.Get(ctx, id)
}

func (self *loggingService) GetByEmail(ctx context.Context, email string) (identity *Identity, err error) {
	defer func(begin time.Time) {
		if err != nil {
			self.logger.Error(
				"Failed to get identity by email.",
				zap.String("email", email),
				zap.Int64("took", time.Since(begin).Milliseconds()),
				zap.String("msg", err.Error()),
			)
			return
		}

		self.logger.Debug(
			"Got identity.",
			zap.String("id", identity.ID),
			zap.Int64("took", time.Since(begin).Milliseconds()),
		)
	}(time.Now())

	return self.Service.GetByEmail(ctx, email)
}
