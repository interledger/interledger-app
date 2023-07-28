package admin

import (
	"context"

	pb "gitlab.com/fynbos/proto/backend/admin/v1"
)

func (s *AdminRpcService) ListLinkedAccounts(ctx context.Context, req *pb.ListLinkedAccountsRequest) (*pb.ListLinkedAccountsResponse, error) {
	lal, err := s.b.LinkedAccounts().ListByWalletId(ctx, req.WalletID)
	if err != nil {
		return nil, err
	}

	resp := make([]*pb.LinkedAccount, len(lal))
	for i, la := range lal {
		resp[i] = &pb.LinkedAccount{
			Id:         la.ID,
			WalletID:   la.WalletID,
			Name:       la.Name,
			Nickname:   la.Nickname,
			Mask:       la.Mask,
			Provider:   la.Provider,
			ProviderID: la.ProviderID,
			Type:       la.Type,
		}
	}

	return &pb.ListLinkedAccountsResponse{Accounts: resp}, err
}

// create an rpc method for getting a linked account by id
func (s *AdminRpcService) GetLinkedAccount(ctx context.Context, req *pb.GetLinkedAccountRequest) (*pb.LinkedAccount, error) {
	la, err := s.b.LinkedAccounts().Get(ctx, req.Id)
	if err != nil {
		return nil, toGRPCError(err)
	}

	return &pb.LinkedAccount{
		Id:         la.ID,
		WalletID:   la.WalletID,
		Name:       la.Name,
		Nickname:   la.Nickname,
		Mask:       la.Mask,
		Provider:   la.Provider,
		ProviderID: la.ProviderID,
		Type:       la.Type,
	}, nil
}
