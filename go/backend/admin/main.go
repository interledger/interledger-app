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
}

func NewServer(args *ServerArgs) (*grpc.Server, error) {
	validator := validator.New()
	if err := validator.Struct(args); err != nil {
		return nil, fmt.Errorf("%w %s", ErrInvalidArgument, err)
	}

	server := grpc.NewServer(
		args.Us.MakeUnaryInterceptors(),
	)
	backendv1.RegisterBackendServiceServer(server, &rpcServer{
		validator: validator,
		as:        args.As,
		is:        args.Is,
		us:        args.Us,
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
}

func (s *rpcServer) GetUserAccountByEmail(
	ctx context.Context,
	req *backendv1.GetUserAccountByEmailRequest,
) (*backendv1.Account, error) {
	_, err := s.us.ForContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "Unauthenticated.")
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
