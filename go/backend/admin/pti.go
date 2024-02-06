package admin

import (
	"context"

	"gitlab.com/fynbos/backend/currency"
	"gitlab.com/fynbos/backend/linkedaccounts"
	"gitlab.com/fynbos/backend/providers/pti"
	adminv1 "gitlab.com/fynbos/proto/backend/admin/v1"
	_ "google.golang.org/protobuf/types/known/emptypb"
)

func (s *AdminRpcService) EnablePTIBalance(
	ctx context.Context, request *adminv1.EnablePTIBalanceRequest,
) (*adminv1.Empty, error) {
	walletId := request.GetWalletId()

	w, err := s.b.Wallets().Get(ctx, walletId)
	if err != nil {
		return nil, toGRPCError(err)
	}

	c := currency.ParseCurrency(currency.USD.String())

	lal, err := s.b.LinkedAccounts().ListByWalletId(ctx, w.ID)
	if err != nil {
		return nil, toGRPCError(err)
	}

	// Check that the linked account doesn't already exist
	for _, la := range lal {
		if la.Provider == pti.ProviderName && la.Type == pti.AccTypeBalance && la.SendCurrency.String() == c.String() {
			return nil, nil
		}
	}

	await, err := s.b.PTI().CreateWallet(ctx, w.ID, c)
	if err != nil {
		return nil, toGRPCError(err)
	}

	var la linkedaccounts.LinkedAccount
	err = await(ctx, &la)
	if err != nil {
		return nil, toGRPCError(err)
	}

	return nil, nil
}

func (s *AdminRpcService) GetPTIBalance(
	ctx context.Context, request *adminv1.GetPTIBalanceRequest,
) (*adminv1.GetPTIBalanceResponse, error) {
	walletId := request.GetWalletId()
	c := currency.ParseCurrency(currency.USD.String())

	w, err := s.b.Wallets().Get(ctx, walletId)
	if err != nil {
		return nil, toGRPCError(err)
	}

	lal, err := s.b.LinkedAccounts().ListByWalletId(ctx, w.ID)
	if err != nil {
		return nil, toGRPCError(err)
	}

	// Check that the linked account doesn't already exist
	var ptiLa *linkedaccounts.LinkedAccount
	for _, la := range lal {
		if la.Provider == pti.ProviderName && la.Type == pti.AccTypeBalance && la.SendCurrency.String() == c.String() {
			ptiLa = &la
			break
		}
	}

	if ptiLa == nil {
		return nil, NotFoundError("balance not found")
	}

	b, err := s.b.PTI().GetBalance(ctx, ptiLa.ID)
	if err != nil {
		return nil, toGRPCError(err)
	}

	return &adminv1.GetPTIBalanceResponse{
		Balance:   b.Total.ToAdminPB(),
		Available: b.Available.ToAdminPB(),
	}, nil
}
