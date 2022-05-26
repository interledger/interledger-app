package grpc

import (
	"context"

	"gitlab.com/fynbos/backend/fundingsources"
	backendv1 "gitlab.com/fynbos/proto/backend/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *rpcService) GetBankAccountWidget(
	ctx context.Context,
	req *backendv1.GetBankAccountWidgetRequest,
) (*backendv1.GetBankAccountWidgetResponse, error) {
	user, err := s.userService.ForContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "Unauthenticated.")
	}

	acc, err := s.accountsService.GetByIdentityID(ctx, user.ID)
	if err != nil {
		return nil, status.Error(codes.Internal, "Unable to get account.")
	}

	url, err := s.fundingSourceService.GetMxConnectWidget(ctx, acc.ID, user.ID)
	if err != nil {
		return nil, status.Error(codes.Internal, "Unable to get widget.")
	}

	return &backendv1.GetBankAccountWidgetResponse{
		Url: url,
	}, nil
}

func (s *rpcService) CreateBankAccount(
	ctx context.Context,
	req *backendv1.CreateBankAccountRequest,
) (*backendv1.FundingSource, error) {
	user, err := s.userService.ForContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "Unauthenticated.")
	}

	acc, err := s.accountsService.GetByIdentityID(ctx, user.ID)
	if err != nil {
		return nil, status.Error(codes.Internal, "Unable to get account.")
	}

	fs, err := s.fundingSourceService.CreateMxBankAccount(ctx, &fundingsources.CreateMxBankAccountArgs{
		AccountID:    acc.ID,
		IdentityID:   user.ID,
		MxUserGuid:   req.GetUserGuid(),
		MxMemberGuid: req.GetMemberGuid(),
		Name:         req.GetName(),
	})
	if err != nil {
		return nil, status.Error(codes.Internal, "Unable to create bank account.")
	}

	return &backendv1.FundingSource{
		Id:    fs.ID,
		State: fs.VerificationState,
		Name:  fs.Name,
		Mask:  fs.Mask,
	}, nil
}
