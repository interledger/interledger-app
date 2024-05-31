package admin

import (
	"context"

	"gitlab.com/fynbos/backend/linkedaccounts"
	"gitlab.com/fynbos/backend/providers/gatehub"
	pb "gitlab.com/fynbos/proto/backend/admin/v1"
)

func (s *AdminRpcService) CreateGatehubUser(
	ctx context.Context, req *pb.CreateGatehubUserRequest,
) (*pb.Empty, error) {
	await, err := s.b.Gatehub().CreateUser(ctx, req.GetWalletID())
	if err != nil {
		return nil, toGRPCError(err)
	}

	err = await(ctx, nil)
	if err != nil {
		return nil, toGRPCError(err)
	}

	return &pb.Empty{}, nil
}

func (s *AdminRpcService) GetGatehubBalance(
	ctx context.Context, req *pb.GetGatehubBalanceRequest,
) (*pb.GetGatehubBalanceResponse, error) {
	balanceLas, err := s.b.LinkedAccounts().ListBalances(ctx, req.GetWalletID())
	if err != nil {
		return nil, toGRPCError(err)
	}

	var gatehubLa *linkedaccounts.LinkedAccount
	for _, balance := range balanceLas {
		if balance.Provider == gatehub.ProviderName && balance.Type == gatehub.AccTypeBalance {
			gatehubLa = &balance
			break
		}
	}
	if gatehubLa == nil {
		return nil, NotFoundError("balance not found")
	}

	bal, err := s.b.Gatehub().GetBalance(ctx, gatehubLa.ID)
	if err != nil {
		return nil, toGRPCError(err)
	}

	return &pb.GetGatehubBalanceResponse{
		Balance:   bal.Total.ToAdminPB(),
		Available: bal.Available.ToAdminPB(),
	}, nil
}
