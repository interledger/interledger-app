package grpc

import (
	"context"

	backendv1 "gitlab.com/fynbos/proto/backend/v1"
)

func (r rpcService) ListBanks(ctx context.Context, args *backendv1.Empty) (*backendv1.ListBanksResponse, error) {
	_, err := r.b.Users().UserForContext(ctx)
	if err != nil {
		return nil, ForbiddenError("Unauthenticated.")
	}
	wallet, err := r.b.Users().WalletForContext(ctx)
	if err != nil {
		return nil, ForbiddenError("Unauthenticated.")
	}

	kyc, err := r.b.KYC().GetIndividualDetails(ctx, wallet.ID)
	if err != nil {
		return nil, toGRPCError(err)
	}

	machnetBanks, err := r.b.Machnet().GetBanks(ctx, kyc.CountryCode)
	if err != nil {
		return nil, toGRPCError(err)
	}

	banks := make([]*backendv1.Bank, len(machnetBanks))
	for i, machnetBank := range machnetBanks {
		branches := make([]*backendv1.Branch, len(machnetBank.Branches))
		for j, machnetBranch := range machnetBank.Branches {
			branches[j] = &backendv1.Branch{
				Id:   machnetBranch.ID,
				Name: machnetBranch.Name,
			}
		}
		banks[i] = &backendv1.Bank{
			Id:       machnetBank.ID,
			Name:     machnetBank.Name,
			Branches: branches,
		}
	}

	return &backendv1.ListBanksResponse{
		Banks: banks,
	}, nil
}

func (r rpcService) CreateReceiveBankAccount(ctx context.Context, args *backendv1.CreateReceiveBankAccountRequest) (*backendv1.LinkedAccount, error) {
	panic("not implemented")
}
