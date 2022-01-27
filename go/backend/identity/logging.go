package identity

import (
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

func (self *loggingService) Create(args CreateArgs) (identity *Identity, err error) {
	defer func(begin time.Time) {
		if err != nil {
			self.logger.Error(
				"Failed to create identity.",
				zap.Int64("took", time.Since(begin).Milliseconds()),
				zap.String("msg", err.Error()),
			)
			return
		}

		self.logger.Debug(
			"Created identity.",
			zap.Int64("took", time.Since(begin).Milliseconds()),
		)
	}(time.Now())

	return self.Service.Create(args)
}

func (self *loggingService) Get(id string) (identity *Identity, err error) {
	defer func(begin time.Time) {
		if err != nil {
			self.logger.Error(
				"Failed to get identity.",
				zap.String("id", id),
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

	return self.Service.Get(id)
}
