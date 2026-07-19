package admin

import (
	"context"

	"github.com/interledger/interledger-app/go/backend/currency"
	"github.com/interledger/interledger-app/go/backend/linkedaccounts"
	"github.com/interledger/interledger-app/go/backend/providers/pti"
	adminv1 "github.com/interledger/interledger-app/go/proto/backend/admin/v1"
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
		if la.Provider == pti.ProviderName && la.Type == pti.AccTypeBalance && la.SendCurrency.String() == c.String() && la.DeletedAt.Time.IsZero() {
			return &adminv1.Empty{}, nil
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

	return &adminv1.Empty{}, nil
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
		if la.Provider == pti.ProviderName && la.Type == pti.AccTypeBalance && la.SendCurrency.String() == c.String() && la.DeletedAt.Time.IsZero() {
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
