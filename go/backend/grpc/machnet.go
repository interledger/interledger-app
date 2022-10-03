package grpc

import (
	"context"

	backendv1 "gitlab.com/fynbos/proto/backend/v1"
)

func (r rpcService) GetMachnetWidgetToken(
	ctx context.Context, req *backendv1.Empty,
) (*backendv1.MachnetWidgetToken, error) {
	_, err := r.b.Users().UserForContext(ctx)
	if err != nil {
		return nil, ForbiddenError("Unauthenticated.")
	}

	wallet, err := r.b.Users().WalletForContext(ctx)
	if err != nil {
		return nil, ForbiddenError("Unauthenticated.")
	}

	token, err := r.b.Machnet().GetWidgetToken(ctx, wallet.ID)
	if err != nil {
		return nil, toGRPCError(err)
	}

	return &backendv1.MachnetWidgetToken{
		Value:            token.Value,
		ExpiresInMinutes: int64(token.ExpiresInMinutes),
	}, nil
}
