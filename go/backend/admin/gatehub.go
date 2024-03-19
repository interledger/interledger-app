package admin

import (
	"context"

	pb "gitlab.com/fynbos/proto/backend/admin/v1"
)

func (s *AdminRpcService) CreateGatehubUser(
	ctx context.Context, req *pb.CreateGatehubUserRequest,
) (*pb.Empty, error) {
	_, err := s.b.Gatehub().CreateUser(ctx, req.GetWalletID())
	if err != nil {
		return nil, toGRPCError(err)
	}

	return &pb.Empty{}, nil
}
