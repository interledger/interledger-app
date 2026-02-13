package grpc

import (
	"context"
	pb "gitlab.com/fynbos/proto/backend/v1"
)

func (s *rpcService) CreateSlackAuthURL(ctx context.Context, _ *pb.Empty) (*pb.CreateSlackAuthURLResponse, error) {
	_, err := s.b.Users().UserForContext(ctx)
	if err != nil {
		return nil, UnauthenticatedError("Unauthenticated.")
	}

	wallet, err := s.b.Wallets().ForContext(ctx)
	if err != nil {
		return nil, UnauthenticatedError("Unauthenticated.")
	}

	url, err := s.b.Slack().CreateAuthURL(ctx, wallet.ID)
	if err != nil {
		return nil, toGRPCError(err)
	}

	return &pb.CreateSlackAuthURLResponse{Url: url}, nil
}
