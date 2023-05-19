package grpc

import (
	"context"
	"errors"
	"gitlab.com/fynbos/backend/user"
	pb "gitlab.com/fynbos/proto/backend/v1"
)

func (s *rpcService) GetCurrentWallet(ctx context.Context, req *pb.Empty) (*pb.GetCurrentWalletResponse, error) {
	_, err := s.b.Users().UserForContext(ctx)
	if err != nil && !errors.Is(err, user.ErrNoUserFound) {
		return nil, ForbiddenError("Unauthenticated.")
	}

	w, err := s.b.Users().WalletForContext(ctx)

	if w == nil {
		return nil, NotFoundError("wallet not found")
	}

	return &pb.GetCurrentWalletResponse{
		Id: w.ID,
	}, toGRPCError(err)
}

func (s *rpcService) SetWalletName(ctx context.Context, req *pb.SetWalletNameRequest) (*pb.Empty, error) {
	_, err := s.b.Users().UserForContext(ctx)
	if err != nil && !errors.Is(err, user.ErrNoUserFound) {
		return nil, UnauthenticatedError("Unauthenticated.")
	}

	w, err := s.b.Users().WalletForContext(ctx)
	if err != nil && !errors.Is(err, user.ErrNoUserFound) {
		return nil, ForbiddenError("Unauthenticated.")
	}

	err = s.b.Users().SetWalletName(ctx, w.ID, req.Name)

	return &pb.Empty{}, toGRPCError(err)
}

func (s *rpcService) GetPublicWalletDetails(ctx context.Context, req *pb.GetPublicWalletDetailsRequest) (*pb.GetPublicWalletDetailsResponse, error) {
	w, err := s.b.Users().GetWallet(ctx, req.GetId())

	return &pb.GetPublicWalletDetailsResponse{
		Id:         w.ID,
		PublicName: w.Name,
	}, toGRPCError(err)
}
