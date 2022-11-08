package grpc

import (
	"context"

	"gitlab.com/fynbos/backend/linkedaccounts"
	"gitlab.com/fynbos/backend/providers/machnet"
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

type validateCreateReceiveBankAccount struct {
	BankID        uint32
	BranchID      uint32
	AccountType   string `validate:"oneof=CHECKING SAVINGS"`
	AccountNumber string `validate:"required"`
	Name          string `validate:"required"`
}

func (r rpcService) CreateReceiveBankAccount(
	ctx context.Context, req *backendv1.CreateReceiveBankAccountRequest,
) (*backendv1.LinkedAccount, error) {
	_, err := r.b.Users().UserForContext(ctx)
	if err != nil {
		return nil, ForbiddenError("Unauthenticated.")
	}
	wallet, err := r.b.Users().WalletForContext(ctx)
	if err != nil {
		return nil, ForbiddenError("Unauthenticated.")
	}

	if err := r.b.Validator().StructCtx(ctx, validateCreateReceiveBankAccount{
		BankID:        req.GetBankId(),
		BranchID:      req.GetBranchId(),
		AccountType:   req.GetAccountType(),
		AccountNumber: req.GetAccountNumber(),
		Name:          req.GetName(),
	}); err != nil {
		return nil, toGRPCError(err)
	}

	bankaccount, err := r.b.Machnet().CreateReceiveBankAccount(ctx, machnet.CreateReceiveBankAccountArgs{
		WalletID:      wallet.ID,
		AccountNumber: req.GetAccountNumber(),
		BankID:        req.GetBankId(),
		BranchID:      req.GetBranchId(),
		Name:          req.GetName(),
		//TODO: otp
	})
	if err != nil {
		return nil, toGRPCError(err)
	}

	linkedaccount, err := r.b.LinkedAccounts().GetByProviderID(ctx, linkedaccounts.GetByProviderIDArgs{
		Provider:   machnet.ProviderName,
		ProviderID: bankaccount.ID,
		Type:       machnet.TypeReceiveBankAccount,
		WalletID:   wallet.ID,
	})
	if err != nil {
		return nil, toGRPCError(err)
	}

	return &backendv1.LinkedAccount{
		Id:   linkedaccount.ID,
		Type: machnet.TypeReceiveBankAccount,
		Name: linkedaccount.Name,
		Mask: linkedaccount.Mask,
	}, nil
}
