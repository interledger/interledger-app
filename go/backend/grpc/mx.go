package grpc

import (
	"context"

	"gitlab.com/fynbos/backend/providers/mx"
	backendv1 "gitlab.com/fynbos/proto/backend/v1"
)

func (s *rpcService) GetMXWidget(
	ctx context.Context, req *backendv1.Empty,
) (*backendv1.MXWidgetResponse, error) {
	_, err := s.b.Users().UserForContext(ctx)
	if err != nil {
		return nil, UnauthenticatedError("Unauthenticated.")
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

func (s *rpcService) CreateMXBankAccounts(
	ctx context.Context, req *backendv1.CreateMXBankAccountsRequest,
) (*backendv1.CreateMXBankAccountsResponse, error) {
	_, err := s.b.Users().UserForContext(ctx)
	if err != nil {
		return nil, UnauthenticatedError("Unauthenticated.")
	}

	wallet, err := s.b.Users().WalletForContext(ctx)
	if err != nil {
		return nil, ForbiddenError("Unauthenticated.")
	}

	las, err := s.b.MX().CreateBankAccounts(ctx, mx.CreateBankAccountsArgs{
		WalletID:    wallet.ID,
		SessionGuid: req.GetSessionGuid(),
		MemberGuid:  req.GetMemberGuid(),
		UserGuid:    req.GetUserGuid(),
	})
	if err != nil {
		return nil, toGRPCError(err)
	}

	ret := make([]*backendv1.LinkedAccount, len(las))
	for i, la := range las {
		ret[i] = &backendv1.LinkedAccount{
			Id:       la.ID,
			Type:     la.Type,
			Name:     la.Name,
			Mask:     la.Mask,
			Nickname: la.Nickname,
		}
	}

	return &backendv1.CreateMXBankAccountsResponse{
		LinkedAccounts: ret,
	}, nil
}
