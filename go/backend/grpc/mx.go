package grpc

import (
	"context"

	"gitlab.com/fynbos/backend/providers/mx"
	backendv1 "gitlab.com/fynbos/proto/backend/v1"
)

func (s *rpcService) GetBankAccountWidget(
	ctx context.Context,
	req *backendv1.GetBankAccountWidgetRequest,
) (*backendv1.GetBankAccountWidgetResponse, error) {
	user, err := s.userService.ForContext(ctx)
	if err != nil {
		return nil, ForbiddenError("Unauthenticated.")
	}

	acc, err := s.accountsService.GetByIdentityID(ctx, user.ID)
	if err != nil {
		return nil, InternalError("Unable to get account.")
	}

	url, err := s.mxProvider.GetConnectWidget(ctx, acc.ID, user.ID)
	if err != nil {
		return nil, InternalError("Unable to get widget.")
	}

	return &backendv1.GetBankAccountWidgetResponse{
		Url: url,
	}, nil
}

func (s *rpcService) InitiateCreateBankAccount(
	ctx context.Context,
	req *backendv1.InitiateCreateBankAccountRequest,
) (*backendv1.InitiateCreateBankAccountResponse, error) {
	user, err := s.userService.ForContext(ctx)
	if err != nil {
		return nil, ForbiddenError("Unauthenticated.")
	}

	acc, err := s.accountsService.GetByIdentityID(ctx, user.ID)
	if err != nil {
		return nil, InternalError("Unable to get account.")
	}

	workflowUuid, err := s.mxProvider.InitiateCreateAccount(ctx, &mx.InitiateCreateAccountArgs{
		AccountID:         acc.ID,
		IdentityID:        user.ID,
		UserGuid:          req.GetUserGuid(),
		MemberGuid:        req.GetMemberGuid(),
		FundingsourceName: req.GetName(),
	})
	if err != nil {
		return nil, InternalError("Unable to create bank account.")
	}

	return &backendv1.InitiateCreateBankAccountResponse{
		Reference: workflowUuid,
	}, nil
}
