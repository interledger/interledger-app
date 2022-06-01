package grpc

import (
	"errors"
	"fmt"

	"github.com/go-playground/validator/v10"
	"gitlab.com/fynbos/backend/accounts"
	_admin "gitlab.com/fynbos/backend/admin"
	"gitlab.com/fynbos/backend/admin/auth"
	"gitlab.com/fynbos/backend/healthcheck"
	"gitlab.com/fynbos/backend/identity"
	"gitlab.com/fynbos/backend/onboarding"
	"gitlab.com/fynbos/backend/providers/unit"
	"gitlab.com/fynbos/backend/user"
	backendv1 "gitlab.com/fynbos/proto/backend/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
)

var (
	ErrInvalidArgument = errors.New("grpc: invalid argument")
	ErrInternal        = errors.New("grpc: internal error")
)

type ServerArgs struct {
	HealthCheckService healthcheck.Service `validate:"required"`
	IdentityService    identity.Service    `validate:"required"`
	AccountsService    accounts.Service    `validate:"required"`
	AdminAuthService   auth.Service        `validate:"required"`
	UserService        user.Service        `validate:"required"`
	UnitProvider       unit.Service        `validate:"required"`
	OnboardingService  onboarding.Service  `validate:"required"`
}

const rpcServiceName = "public"

type rpcService struct {
	backendv1.UnimplementedBackendServiceServer
	validator         *validator.Validate
	accountsService   accounts.Service
	identityService   identity.Service
	userService       user.Service
	unitProvider      unit.Service
	onboardingService onboarding.Service
	name              string
}

func (s *rpcService) GetName() string {
	return s.name
}

func NewServer(args *ServerArgs) (*grpc.Server, error) {
	v := validator.New()
	if err := v.Struct(args); err != nil {
		return nil, fmt.Errorf("%w %s", ErrInvalidArgument, err)
	}

	server := grpc.NewServer(
		args.AdminAuthService.MakeUnaryInterceptors(_admin.AdminRpcServiceName),
		user.MakeUnaryInterceptor(args.UserService, rpcServiceName),
	)
	backendv1.RegisterBackendServiceServer(server, &rpcService{
		validator:         v,
		accountsService:   args.AccountsService,
		identityService:   args.IdentityService,
		userService:       args.UserService,
		unitProvider:      args.UnitProvider,
		onboardingService: args.OnboardingService,
		name:              rpcServiceName,
	})
	backendv1.RegisterBackendAdminServiceServer(server, &_admin.AdminRpcService{
		Validator:       v,
		AccountsService: args.AccountsService,
		IdentityService: args.IdentityService,
		AuthService:     args.AdminAuthService,
		UnitService:     args.UnitProvider,
		Name:            _admin.AdminRpcServiceName,
	})
	grpc_health_v1.RegisterHealthServer(server, args.HealthCheckService)
	reflection.Register(server)
	return server, nil
}
