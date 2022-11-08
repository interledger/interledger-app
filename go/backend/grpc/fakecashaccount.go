package grpc

import (
	"context"

	"gitlab.com/fynbos/backend/linkedaccounts"
	"gitlab.com/fynbos/backend/providers/fakecash"
	backendv1 "gitlab.com/fynbos/proto/backend/v1"
)

func (s *rpcService) LinkCashAccount(
	ctx context.Context,
	req *backendv1.LinkCashAccountRequest,
) (*backendv1.LinkedAccount, error) {
	_, err := s.b.Users().UserForContext(ctx)
	if err != nil {
		return nil, ForbiddenError("Unauthenticated.")
	}

	wallet, err := s.b.Users().WalletForContext(ctx)
	if err != nil {
		return nil, ForbiddenError("Unauthenticated.")
	}

	if err = s.b.Validator().VarCtx(ctx, req.GetName(), "required"); err != nil {
		return nil, toGRPCError(err)
	}

	fakecashAccount, err := s.b.FakeCash().Create(ctx, fakecash.CreateArgs{})
	if err != nil {
		return nil, toGRPCError(err)
	}

	linkedAccount, err := s.b.LinkedAccounts().Create(
		ctx,
		&linkedaccounts.CreateArgs{
			WalletID:   wallet.ID,
			Name:       req.GetName(),
			Provider:   "fakecash",
			ProviderID: fakecashAccount.ID,
			Type:       "fakecash",
		},
	)
	if err != nil {
		return nil, toGRPCError(err)
	}

	return &backendv1.LinkedAccount{
		Id:   linkedAccount.ID,
		Name: linkedAccount.Name,
		Type: "fakecash",
	}, nil
}
