package grpc

import (
	"context"
	"errors"

	"github.com/interledger/interledger-app/go/backend/wallets"

	"github.com/interledger/interledger-app/go/backend/user"
	pb "github.com/interledger/interledger-app/go/proto/backend/v1"
)

func (s *rpcService) CreateUserDefaultWallet(ctx context.Context, req *pb.CreateUserDefaultWalletRequest) (*pb.Empty, error) {
	u, err := s.b.Users().UserForContext(ctx)
	if err != nil && !errors.Is(err, user.ErrNoUserFound) {
		return nil, UnauthenticatedError("Unauthenticated.")
	}

	if u != nil && u.ID != req.UserID {
		return nil, ForbiddenError("Unauthenticated.")
	}

	_, err = s.b.Wallets().Create(ctx, wallets.CreateArgs{
		UserID: req.UserID,
	})

	return &pb.Empty{}, toGRPCError(err)
}

func (s *rpcService) UpdateUserPhone(ctx context.Context, req *pb.UpdateUserPhoneRequest) (*pb.Empty, error) {
	u, err := s.b.Users().UserForContext(ctx)
	if err != nil {
		return nil, UnauthenticatedError("Unauthenticated.")
	}

	if err := s.b.Validator().VarCtx(ctx, req.GetPhone(), "required,e164"); err != nil {
		return nil, NewValidationError("phone", "Phone number is invalid.")
	}

	err = s.b.Users().UpdateUserPhone(ctx, u.ID, req.GetPhone())
	if err != nil {
		return nil, toGRPCError(err)
	}

	return &pb.Empty{}, nil
}
