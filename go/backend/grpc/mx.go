package grpc

import (
	"context"

	backendv1 "gitlab.com/fynbos/proto/backend/v1"
)

func (s *rpcService) GetMXWidget(
	ctx context.Context, req *backendv1.Empty,
) (*backendv1.MXWidgetResponse, error) {
	_, err := s.b.Users().UserForContext(ctx)
	if err != nil {
		return nil, ForbiddenError("Unauthenticated.")
	}

	wallet, err := s.b.Users().WalletForContext(ctx)
	if err != nil {
		return nil, ForbiddenError("Unauthenticated.")
	}

	widgetUrl, err := s.b.MX().GetWidget(ctx, wallet.ID)
	if err != nil {
		return nil, toGRPCError(err)
	}

	return &backendv1.MXWidgetResponse{
		Url: widgetUrl,
	}, nil
}
