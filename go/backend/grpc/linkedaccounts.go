package grpc

import (
	"context"

	backendv1 "gitlab.com/fynbos/proto/backend/v1"
)

func (s *rpcService) GetLinkedAccounts(
	ctx context.Context, _ *backendv1.Empty,
) (*backendv1.GetLinkedAccountsResponse, error) {
	_, err := s.b.Users().UserForContext(ctx)
	if err != nil {
		return nil, ForbiddenError("Unauthenticated.")
	}

	wallet, err := s.b.Users().WalletForContext(ctx)
	if err != nil {
		return nil, ForbiddenError("Unauthenticated.")
	}

	linkedAccounts, err := s.b.LinkedAccounts().ListByWalletId(ctx, wallet.ID)
	if err != nil {
		return nil, InternalError("Unable to get linked accounts.")
	}

	ret := make([]*backendv1.LinkedAccount, len(linkedAccounts))
	for i, fs := range linkedAccounts {
		ret[i] = &backendv1.LinkedAccount{
			Id:   fs.ID,
			Name: fs.Name,
			Mask: fs.Mask,
		}
	}

	return &backendv1.GetLinkedAccountsResponse{
		LinkedAccounts: ret,
	}, nil
}
