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
	"gitlab.com/fynbos/backend/providers/unit"
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
	Hs healthcheck.Service `validate:"required"`
	Is identity.Service    `validate:"required"`
	As accounts.Service    `validate:"required"`
	Us auth.Service        `validate:"required"`
	Up unit.Service        `validate:"required"`
}

type rpcServer struct {
	backendv1.UnimplementedBackendServiceServer
	validator *validator.Validate
	as        accounts.Service
	is        identity.Service
	us        auth.Service
	up        unit.Service
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
	backendv1.RegisterBackendAdminServiceServer(server, &_admin.AdminRpcServer{
		Validator:       v,
		AccountsService: args.As,
		IdentityService: args.Is,
		AuthService:     args.Us,
		UnitService:     args.Up,
	})
	grpc_health_v1.RegisterHealthServer(server, args.Hs)
	reflection.Register(server)
	return server, nil
}
