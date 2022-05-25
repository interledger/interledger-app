package grpc

import (
	"context"

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
