package grpc

import (
	"errors"
	"fmt"

	"gitlab.com/fynbos/backend/payments"

	"gitlab.com/fynbos/backend/supporttickets"

	"github.com/go-playground/validator/v10"
	"gitlab.com/fynbos/backend/accounts"
	_admin "gitlab.com/fynbos/backend/admin"
	"gitlab.com/fynbos/backend/admin/auth"
	"gitlab.com/fynbos/backend/agreements"
	"gitlab.com/fynbos/backend/deposits"
	"gitlab.com/fynbos/backend/fundingsources"
	"gitlab.com/fynbos/backend/healthcheck"
	"gitlab.com/fynbos/backend/identity"
	"gitlab.com/fynbos/backend/onboarding"
	"gitlab.com/fynbos/backend/providers/mx"
	"gitlab.com/fynbos/backend/providers/rafiki"
	"gitlab.com/fynbos/backend/providers/unit"
	"gitlab.com/fynbos/backend/twilio"
	"gitlab.com/fynbos/backend/user"
	"gitlab.com/fynbos/backend/waitlist"
	backendv1 "gitlab.com/fynbos/proto/backend/v1"
	"go.temporal.io/sdk/client"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
)

var (
	ErrInvalidArgument = errors.New("grpc: invalid argument")
	ErrInternal        = errors.New("grpc: internal error")
)

type ServerArgs struct {
	HealthCheckService   healthcheck.Service   `validate:"required"`
	IdentityService      identity.Client       `validate:"required"`
	AccountsService      accounts.Client       `validate:"required"`
	AgreementsService    agreements.Service    `validate:"required"`
	AdminAuthService     auth.Service          `validate:"required"`
	UserService          user.Service          `validate:"required"`
	UnitProvider         unit.Client           `validate:"required"`
	FundingSourceService fundingsources.Client `validate:"required"`
	OnboardingService    onboarding.Client     `validate:"required"`
	MxProvider           mx.Client             `validate:"required"`
	RafikiProvider       rafiki.Service        `validate:"required"`
	DepositService       deposits.Service      `validate:"required"`
	TwilioService        twilio.Service        `validate:"required"`
	WaitlistClient       waitlist.Client       `validate:"required"`
	Temporal             client.Client         `validate:"required"`
	TicketClient         supporttickets.Client `validate:"required"`
	PaymentsClient       payments.Client       `validate:"required"`
}

type rpcService struct {
	validator            *validator.Validate
	accountsService      accounts.Client
	agreementsService    agreements.Service
	identityService      identity.Client
	userService          user.Service
	unitProvider         unit.Client
	onboardingService    onboarding.Client
	fundingSourceService fundingsources.Client
	mxProvider           mx.Client
	rafikiProvider       rafiki.Service
	depositService       deposits.Service
	twilioService        twilio.Service
	waitlistClient       waitlist.Client
	ticketsClient        supporttickets.Client
	paymentsClient       payments.Client
}

func NewServer(args *ServerArgs) (*grpc.Server, error) {
	v := validator.New()
	if err := v.Struct(args); err != nil {
		return nil, fmt.Errorf("%w %s", ErrInvalidArgument, err)
	}

	server := grpc.NewServer(
		args.AdminAuthService.MakeUnaryInterceptors(),
		user.MakeUnaryInterceptor(args.UserService),
	)
	backendv1.RegisterBackendServiceServer(server, &rpcService{
		validator:            v,
		accountsService:      args.AccountsService,
		agreementsService:    args.AgreementsService,
		identityService:      args.IdentityService,
		userService:          args.UserService,
		unitProvider:         args.UnitProvider,
		onboardingService:    args.OnboardingService,
		fundingSourceService: args.FundingSourceService,
		mxProvider:           args.MxProvider,
		rafikiProvider:       args.RafikiProvider,
		depositService:       args.DepositService,
		twilioService:        args.TwilioService,
		waitlistClient:       args.WaitlistClient,
		ticketsClient:        args.TicketClient,
		paymentsClient:       args.PaymentsClient,
	})
	backendv1.RegisterBackendAdminServiceServer(server, &_admin.AdminRpcService{
		Validator:       v,
		AccountsService: args.AccountsService,
		IdentityService: args.IdentityService,
		AuthService:     args.AdminAuthService,
		UnitService:     args.UnitProvider,
		Temporal:        args.Temporal,
	})
	grpc_health_v1.RegisterHealthServer(server, args.HealthCheckService)
	reflection.Register(server)
	return server, nil
}
