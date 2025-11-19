package grpc

import (
	"context"
	"errors"

	"gitlab.com/fynbos/backend/wallets"

	"gitlab.com/fynbos/backend/user"
	pb "gitlab.com/fynbos/proto/backend/v1"
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

func (s *rpcService) SendAccountVerificationEmail(ctx context.Context, req *pb.Empty) (*pb.Empty, error) {
	u, err := s.b.Users().UserForContext(ctx)
	if err != nil && !errors.Is(err, user.ErrNoUserFound) {
		return nil, UnauthenticatedError("Unauthenticated.")
	}

	err = s.b.Email().SendAccountVerificationEmail(ctx, u.Email)

	return &pb.Empty{}, toGRPCError(err)
}

