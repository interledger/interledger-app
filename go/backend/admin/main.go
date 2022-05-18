package admin

import (
	"context"
	"errors"
	"fmt"

	"github.com/go-playground/validator/v10"
	"gitlab.com/fynbos/backend/accounts"
	"gitlab.com/fynbos/backend/admin/auth"
	"gitlab.com/fynbos/backend/healthcheck"
	"gitlab.com/fynbos/backend/identity"
	"gitlab.com/fynbos/backend/providers/unit"
	"gitlab.com/fynbos/proto/backend/v1"
	backendv1 "gitlab.com/fynbos/proto/backend/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
)

var (
	ErrInvalidArgument = errors.New("admin grpc: invalid argument.")
	ErrInternal        = errors.New("admin grpc: internal error.")
)

type ServerArgs struct {
	Hs healthcheck.Service `validate:"required"`
	Is identity.Service    `validate:"required"`
	As accounts.Service    `validate:"required"`
	Us auth.Service        `validate:"required"`
	Up unit.Service        `validate:"required"`
}

func NewServer(args *ServerArgs) (*grpc.Server, error) {
	v := validator.New()
	if err := v.Struct(args); err != nil {
		return nil, fmt.Errorf("%w %s", ErrInvalidArgument, err)
	}

	server := grpc.NewServer(
		args.Us.MakeUnaryInterceptors(),
	)
	backendv1.RegisterBackendServiceServer(server, &rpcServer{
		validator: v,
		as:        args.As,
		is:        args.Is,
		us:        args.Us,
		up:        args.Up,
	})
	grpc_health_v1.RegisterHealthServer(server, args.Hs)
	reflection.Register(server)
	return server, nil
}

type rpcServer struct {
	backendv1.UnimplementedBackendServiceServer
	validator *validator.Validate
	as        accounts.Service
	is        identity.Service
	us        auth.Service
	up        unit.Service
}

func (s *rpcServer) GetUserAccountByEmail(
	ctx context.Context,
	req *backendv1.GetUserAccountByEmailRequest,
) (*backendv1.Account, error) {
	adminUser, err := s.us.ForContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "Unauthenticated.")
	}
	isAdmin := authorizeAdmin(adminUser.Email)
	if !isAdmin {
		return nil, status.Error(codes.PermissionDenied, "Forbidden.")
	}

	id, err := s.is.GetByEmail(ctx, req.GetEmail())
	if err != nil {
		switch {
		case errors.Is(err, identity.ErrNotFound):
			return nil, status.Error(codes.NotFound, err.Error())
		default:
			return nil, status.Error(codes.Internal, err.Error())
		}
	}

	acc, err := s.as.GetByIdentityID(ctx, id.ID)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &backendv1.Account{
		Id:              acc.ID,
		Balance:         acc.AvailableBalance,
		DebitsReserved:  acc.DebitsReserved,
		DebitsAccepted:  acc.DebitsAccepted,
		CreditsReserved: acc.CreditsReserved,
		CreditsAccepted: acc.CreditsAccepted,
	}, nil
}

func (s *rpcServer) GetUnitCustomerByAccountID(
	ctx context.Context,
	req *backendv1.GetUnitCustomerByAccountRequest,
) (*backend.UnitCustomer, error) {
	adminUser, err := s.us.ForContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "Unauthenticated.")
	}
	isAdmin := authorizeAdmin(adminUser.Email)
	if !isAdmin {
		return nil, status.Error(codes.PermissionDenied, "Forbidden.")
	}

	customer, err := s.up.GetCustomerByAccountID(ctx, req.GetAccountId())
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &backendv1.UnitCustomer{
		Id:        customer.ID,
		AccountId: customer.AccountID,
		Type:      customer.Type,
	}, nil
}

func authorizeAdmin(email string) bool {
	emails := [...]string{
		"don@fynbos.dev",
		"matt@fynbos.dev",
		"cairin@fynbos.dev",
		"adrian@fynbos.dev",
	}
	for _, e := range emails {
		if e == email {
			return true
		}
	}
	return false
}
