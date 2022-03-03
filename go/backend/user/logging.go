package user

import (
	"context"
	"errors"
	"net/http"
	"time"

	"go.uber.org/zap"
)

type loggingService struct {
	logger *zap.Logger
	Service
}

func NewLoggingService(us Service, logger *zap.Logger) Service {
	childLogger := logger.With(zap.String("service", "user"))
	return &loggingService{childLogger, us}
}

func (self *loggingService) GetUser(r http.Request) (usr *User, err error) {
	defer func(begin time.Time) {
		if err != nil {
			if !errors.Is(err, ErrNoUserFound) {
				self.logger.Error(
					"failed to parse user cookie",
				)
			}
		}
	}(time.Now())

	return self.Service.GetUser(r)
}

func (self *loggingService) ForContext(ctx context.Context) (usr *User, err error) {
	defer func(begin time.Time) {
		if err != nil {
			self.logger.Info(
				"no user in context",
			)
		}
	}(time.Now())

	return self.Service.ForContext(ctx)
}
