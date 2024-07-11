package grpc

import (
	"context"
	"errors"
	"sort"

	"gitlab.com/fynbos/backend/providers/chimoney"
	"gitlab.com/fynbos/backend/providers/gatehub"
	"gitlab.com/fynbos/backend/providers/pti"
	"gitlab.com/fynbos/backend/providers/xago"
	"gitlab.com/fynbos/backend/user"
	"gitlab.com/fynbos/proto/backend/v1"
	pb "gitlab.com/fynbos/proto/backend/v1"
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
		if la.Provider == pti.ProviderName && la.Type == pti.AccTypeBalance && la.DeletedAt.Time.IsZero() {
			ptiBalance, err := s.b.PTI().GetBalance(ctx, la.ID)
			if err != nil {
				return nil, toGRPCError(err)
			}

			resp = append(resp, &pb.Balance{
				Balance:          ptiBalance.Available.ToPB(),
				Currency:         la.SendCurrency.String(),
				CountryCode:      la.ReceiveCountry.String(),
				LinkedAccount:    la.ID,
				FormattedBalance: ptiBalance.Available.Format(),
			})
		}

		if la.Provider == xago.ProviderName && la.Type == xago.AccTypeBalance && la.DeletedAt.Time.IsZero() {
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

		if la.Provider == gatehub.ProviderName && la.Type == gatehub.AccTypeBalance && la.DeletedAt.Time.IsZero() {
			bal, err := s.b.Gatehub().GetBalance(ctx, la.ID)
			if err != nil {
				return nil, toGRPCError(err)
			}

			resp = append(resp, &pb.Balance{
				Balance:          bal.Available.ToPB(),
				Currency:         la.SendCurrency.String(),
				CountryCode:      "EU", // EUR is a special case where we want to show the Euro flag on the frontend
				LinkedAccount:    la.ID,
				FormattedBalance: bal.Available.Format(),
			})
		}

		if la.Provider == chimoney.ProviderName && la.Type == chimoney.AccTypeBalance && la.DeletedAt.Time.IsZero() {
			bal, err := s.b.Chimoney().GetBalance(ctx, la.ID)
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
