package grpc

import (
	"context"

	pb "gitlab.com/fynbos/proto/backend/v1"
)

func (s *rpcService) OnboardGMTUser(ctx context.Context, _ *pb.Empty) (*pb.Empty, error) {
	_, err := s.b.Users().UserForContext(ctx)
	if err != nil {
		return nil, UnauthenticatedError("Unauthenticated.")
	}

	wallet, err := s.b.Wallets().ForContext(ctx)
	if err != nil {
		return nil, ForbiddenError("Unauthenticated.")
	}

	_, err = s.b.GMT().StartUserOnboarding(ctx, wallet.ID)

	return &pb.Empty{}, toGRPCError(err)
}
