package admin

import (
	"context"

	"gitlab.com/fynbos/log"
	pb "gitlab.com/fynbos/proto/backend/admin/v1"
	"go.uber.org/zap"
)

func (s *AdminRpcService) CreateGatehubUser(
	ctx context.Context, req *pb.CreateGatehubUserRequest,
) (*pb.Empty, error) {
	externalUserID, err := s.b.Gatehub().CreateUser(ctx, req.GetWalletID())
	if err != nil {
		return nil, toGRPCError(err)
	}

	log.Info("created gatehub user", zap.String("walletID", req.GetWalletID()), zap.String("externalID", externalUserID))

	err = s.b.Gatehub().SaveUser(ctx, req.GetWalletID(), externalUserID)
	if err != nil {
		return nil, toGRPCError(err)
	}

	return &pb.Empty{}, nil
}
