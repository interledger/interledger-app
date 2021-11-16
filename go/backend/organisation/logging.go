package organisation

import (
	"time"

	"gitlab.com/fynbos/backend/user"
	"go.uber.org/zap"
)

type loggingService struct {
	logger  *zap.Logger
	Service Service
}

func NewLoggingService(orgService Service, logger *zap.Logger) Service {
	childLogger := logger.With(zap.String("service", "organisation"))
	return &loggingService{childLogger, orgService}
}

func (self *loggingService) Create(name string, user user.User) (org *Organisation, err error) {
	defer func(begin time.Time) {
		if err != nil {
			self.logger.Error(
				"Failed to create organisation.",
				zap.Int64("took", time.Since(begin).Milliseconds()),
				zap.String("msg", err.Error()),
			)
		}

		self.logger.Debug(
			"Created organisation.",
			zap.Int64("took", time.Since(begin).Milliseconds()),
		)
	}(time.Now())

	return self.Service.Create(name, user)
}

func (self *loggingService) Get(id string, user user.User) (org *Organisation, err error) {
	defer func(begin time.Time) {
		if err != nil {
			self.logger.Error(
				"Failed to get organisation.",
				zap.String("id", id),
				zap.Int64("took", time.Since(begin).Milliseconds()),
				zap.String("msg", err.Error()),
			)
		}

		self.logger.Debug(
			"Got organisation.",
			zap.String("id", id),
			zap.Int64("took", time.Since(begin).Milliseconds()),
		)
	}(time.Now())

	return self.Service.Get(id, user)
}

func (self loggingService) GetAllOwnedBy(user user.User) (orgs []*Organisation, err error) {
	defer func(begin time.Time) {
		if err != nil {
			self.logger.Error(
				"Failed to get users organisations.",
				zap.String("userID", user.ID),
				zap.Int64("took", time.Since(begin).Milliseconds()),
				zap.String("msg", err.Error()),
			)
		}

		self.logger.Debug(
			"Got user organisations.",
			zap.String("userID", user.ID),
			zap.Int64("took", time.Since(begin).Milliseconds()),
		)
	}(time.Now())

	return self.Service.GetAllOwnedBy(user)
}
