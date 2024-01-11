package grpc

import (
	"context"
	"errors"

	"gitlab.com/fynbos/backend/providers/pti"
	"gitlab.com/fynbos/backend/user"
	"gitlab.com/fynbos/proto/backend/v1"
	pb "gitlab.com/fynbos/proto/backend/v1"
)

func (s *rpcService) GetPtiBalances(ctx context.Context, req *backend.Empty) (*backend.GetPtiBalancesResponse, error) {
	_, err := s.b.Users().UserForContext(ctx)
	if err != nil && !errors.Is(err, user.ErrNoUserFound) {
		return nil, UnauthenticatedError("Unauthenticated.")
	}

	w, err := s.b.Wallets().ForContext(ctx)
	if err != nil && !errors.Is(err, user.ErrNoUserFound) {
		return nil, ForbiddenError("Unauthenticated.")
	}

	lal, err := s.b.LinkedAccounts().ListByWalletId(ctx, w.ID)
	if err != nil {
		return nil, toGRPCError(err)
	}

	var resp []*pb.PtiBalance
	for _, la := range lal {
		if la.Provider != pti.ProviderName || la.Type != pti.AccTypeBalance {
			continue
		}

		ptiWallet, err := s.b.PTI().GetWallet(ctx, la.ID)
		if err != nil {
			return nil, toGRPCError(err)
		}

		resp = append(resp, &pb.PtiBalance{
			Balance:          ptiWallet.Balance.ToPB(),
			Currency:         la.SendCurrency.String(),
			LinkedAccount:    la.ID,
			FormattedBalance: ptiWallet.Balance.Format(),
		})
	}

	return &backend.GetPtiBalancesResponse{
		Balances: resp,
	}, nil
}
