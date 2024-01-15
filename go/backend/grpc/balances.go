package grpc

import (
	"context"
	"errors"
	"gitlab.com/fynbos/backend/providers/pti"
	"gitlab.com/fynbos/backend/providers/xago"
	"gitlab.com/fynbos/backend/user"
	"gitlab.com/fynbos/proto/backend/v1"
	pb "gitlab.com/fynbos/proto/backend/v1"
	"sort"
)

func (s *rpcService) GetBalances(ctx context.Context, req *backend.Empty) (*backend.GetBalancesResponse, error) {
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

	var resp []*pb.Balance
	for _, la := range lal {
		if la.Provider == pti.ProviderName && la.Type == pti.AccTypeBalance {
			ptiWallet, err := s.b.PTI().GetWallet(ctx, la.ID)
			if err != nil {
				return nil, toGRPCError(err)
			}

			resp = append(resp, &pb.Balance{
				Balance:          ptiWallet.Balance.ToPB(),
				Currency:         la.SendCurrency.String(),
				CountryCode:      la.ReceiveCountry.String(),
				LinkedAccount:    la.ID,
				FormattedBalance: ptiWallet.Balance.Format(),
			})
		}

		if la.Provider == xago.ProviderName && la.Type == xago.AccTypeBalance {
			bal, err := s.b.Xago().GetBalance(ctx, la.ID)
			if err != nil {
				return nil, toGRPCError(err)
			}

			resp = append(resp, &pb.Balance{
				Balance:          bal.Available.ToPB(),
				Currency:         la.SendCurrency.String(),
				CountryCode:      la.ReceiveCountry.String(),
				LinkedAccount:    la.ID,
				FormattedBalance: bal.Available.Format(),
			})
		}
	}

	// Put the balances of the wallets default country first
	sort.Slice(resp, func(i, j int) bool {
		return resp[i].CountryCode == w.Country.String()
	})

	return &backend.GetBalancesResponse{
		Balances: resp,
	}, nil
}
