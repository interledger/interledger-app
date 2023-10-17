package admin

import (
	"context"
	"gitlab.com/fynbos/env"
	pb "gitlab.com/fynbos/proto/backend/admin/v1"
)

func (s *AdminRpcService) ClearIdentities(ctx context.Context, req *pb.ClearIdentitiesRequest) (*pb.Empty, error) {
	if env.IsProd() {
		return nil, UnimplementedError("")
	}

	err := s.b.AdminIdentities().Clear(ctx, req.WalletId)
	if err != nil {
		return nil, toGRPCError(err)
	}

	return &pb.Empty{}, nil
}
