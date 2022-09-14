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
	user, err := s.b.Users().ForContext(ctx)
	if err != nil {
		return nil, UnauthenticatedError("no user.")
	}

	acc, err := s.b.Accounts().GetByIdentityID(ctx, user.ID)
	if err != nil {
		return nil, InternalError("Account not found.")
	}

	deposit, err := s.b.Deposits().InitiateDeposit(ctx, &deposits.InitiateDepositArgs{
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

func (s *rpcService) GetDeposit(ctx context.Context,
	req *backendv1.GetDepositRequest,
) (*backendv1.Deposit, error) {
	user, err := s.b.Users().ForContext(ctx)
	if err != nil {
		return nil, UnauthenticatedError("no user.")
	}

	acc, err := s.b.Accounts().GetByIdentityID(ctx, user.ID)
	if err != nil {
		return nil, InternalError("Account not found.")
	}

	deposit, err := s.b.Deposits().Get(ctx, req.GetId())
	if err != nil {
		return nil, InternalError("Failed to get deposit.")
	}
	if deposit.AccountID != acc.ID {
		return nil, InternalError("Failed to get deposit.")
	}

	return &backendv1.Deposit{
		Id:              deposit.ID,
		FundingsourceId: deposit.FundingSourceId,
		Amount:          deposit.Amount,
		State:           deposit.State.String(),
	}, nil
}
