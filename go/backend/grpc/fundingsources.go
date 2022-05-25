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

	return &backendv1.GetBankAccountWidgetResponse{
		Url: user.ID,
	}, nil
}
