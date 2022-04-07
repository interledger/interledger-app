package auth

import (
	"context"
	"time"

	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type loggingService struct {
	logger  *zap.Logger
	Service Service
}

func NewLoggingService(adminUserService Service, logger *zap.Logger) Service {
	childLogger := logger.With(zap.String("service", "admin user"))
	return &loggingService{childLogger, adminUserService}
}

func (s *loggingService) GetAdminUser(ctx context.Context) (user *AdminUser, err error) {
	defer func(begin time.Time) {
		if err != nil {
			s.logger.Error(
				"Failed to get admin user from token in ctx.",
				zap.String("msg", err.Error()),
			)
			return
		}

		s.logger.Debug("Got admin user from token in ctx.", zap.String("email", user.Email))
	}(time.Now())
	return s.Service.GetAdminUser(ctx)
}

func (s *loggingService) ForContext(ctx context.Context) (user *AdminUser, err error) {
	defer func(begin time.Time) {
		if err != nil {
			s.logger.Error(
				"Failed to get admin user from ctx.",
				zap.String("msg", err.Error()),
			)
			return
		}

		s.logger.Debug("Got admin user from ctx.", zap.String("email", user.Email))
	}(time.Now())
	return s.Service.ForContext(ctx)
}

func (s *loggingService) MakeUnaryInterceptors() grpc.ServerOption {
	defer func() {
		s.logger.Debug("Made auth unary interceptors.")
	}()

	return s.Service.MakeUnaryInterceptors()
}
