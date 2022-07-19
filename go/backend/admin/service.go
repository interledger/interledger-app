package admin

import (
	"context"
	"errors"

	"gitlab.com/fynbos/backend/accounts/ops"

	"github.com/go-playground/validator/v10"
	"gitlab.com/fynbos/backend/admin/auth"
	"gitlab.com/fynbos/backend/identity"
	"gitlab.com/fynbos/backend/providers/unit"
	backendv1 "gitlab.com/fynbos/proto/backend/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type AdminRpcService struct {
	backendv1.UnimplementedBackendAdminServiceServer
	Validator       *validator.Validate
	AccountsService ops.Service
	IdentityService identity.Service
	AuthService     auth.Service
	UnitService     unit.Service
}

func (s *AdminRpcService) GetUserAccountByEmail(
	ctx context.Context,
	req *backendv1.GetUserAccountByEmailRequest,
) (*backendv1.Account, error) {
	adminUser, err := s.AuthService.ForContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "Unauthenticated.")
	}
	isAdmin := authorizeAdmin(adminUser.Email)
	if !isAdmin {
		return nil, status.Error(codes.PermissionDenied, "Forbidden.")
	}

	id, err := s.IdentityService.GetByEmail(ctx, req.GetEmail())
	if err != nil {
		switch {
		case errors.Is(err, identity.ErrNotFound):
			return nil, status.Error(codes.NotFound, err.Error())
		default:
			return nil, status.Error(codes.Internal, err.Error())
		}
	}

	acc, err := s.AccountsService.GetByIdentityID(ctx, id.ID)
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

func (s *AdminRpcService) GetUnitCustomerByAccountID(
	ctx context.Context,
	req *backendv1.GetUnitCustomerByAccountRequest,
) (*backendv1.UnitCustomer, error) {
	adminUser, err := s.AuthService.ForContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "Unauthenticated.")
	}
	isAdmin := authorizeAdmin(adminUser.Email)
	if !isAdmin {
		return nil, status.Error(codes.PermissionDenied, "Forbidden.")
	}

	customer, err := s.UnitService.GetCustomerByAccountID(ctx, req.GetAccountId())
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
