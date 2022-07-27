package grpc

import (
	"context"

	"gitlab.com/fynbos/backend/deposits"
	backendv1 "gitlab.com/fynbos/proto/backend/v1"
)

func (s *rpcService) InitiateDeposit(
	ctx context.Context,
	req *backendv1.InitiateDepositRequest,
) (*backendv1.InitiateDepositResponse, error) {
	user, err := s.userService.ForContext(ctx)
	if err != nil {
		return nil, UnauthenticatedError("no user.")
	}

	acc, err := s.accountsService.GetByIdentityID(ctx, user.ID)
	if err != nil {
		return nil, InternalError("Account not found.")
	}

	deposit, err := s.depositService.InitiateDeposit(ctx, &deposits.InitiateDepositArgs{
		IdentityID:      user.ID,
		AccountID:       acc.ID,
		FundingSourceID: req.GetFundingsourceId(),
		Amount:          req.GetAmount(),
	})
	if err != nil {
		return nil, InternalError("Failed to initiate deposit.")
	}

	return &backendv1.InitiateDepositResponse{
		DepositId: deposit.ID,
	}, nil
}
