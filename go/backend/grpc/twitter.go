package grpc

import (
	"context"

	"github.com/google/uuid"
	"gitlab.com/fynbos/backend/twitter"
	pb "gitlab.com/fynbos/proto/backend/v1"
)

func (s *rpcService) CreateAuthURL(
	ctx context.Context,
	_ *pb.Empty,
) {
	state := uuid.NewString()
	url := s.b.Twitter().CreateAuthURL(ctx, &twitter.CreateAuthURLArgs{

	})

	return s.b.Twitter().CreateAuthURL(ctx)
}