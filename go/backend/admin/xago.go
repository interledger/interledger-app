package admin

import (
	"context"

	"github.com/interledger/interledger-app/go/backend/currency"
	"github.com/interledger/interledger-app/go/backend/linkedaccounts"
	"github.com/interledger/interledger-app/go/backend/providers/xago"
	adminv1 "github.com/interledger/interledger-app/go/proto/backend/admin/v1"
	_ "google.golang.org/protobuf/types/known/emptypb"
)

func (s *AdminRpcService) GetWalletXagoBalance(
	ctx context.Context, request *adminv1.GetWalletXagoBalanceRequest,
) (*adminv1.GetWalletXagoBalanceResponse, error) {
	walletId := request.GetWalletId()
	c := currency.ParseCurrency(currency.ZAR.String())

	w, err := s.b.Wallets().Get(ctx, walletId)
	if err != nil {
		return nil, toGRPCError(err)
	}

	lal, err := s.b.LinkedAccounts().ListByWalletId(ctx, w.ID)
	if err != nil {
		return nil, toGRPCError(err)
	}

	// Check that the linked account doesn't already exist
	var xagoLa *linkedaccounts.LinkedAccount
	for _, la := range lal {
		if la.Provider == xago.ProviderName && la.Type == xago.AccTypeBalance && la.SendCurrency.String() == c.String() {
			xagoLa = &la
			break
		}
	}

	if xagoLa == nil {
		return nil, NotFoundError("balance not found")
	}

	b, err := s.b.Xago().GetBalance(ctx, xagoLa.ID)
	if err != nil {
		return nil, toGRPCError(err)
	}

	return &adminv1.GetWalletXagoBalanceResponse{
		Balance:   b.Total.ToAdminPB(),
		Available: b.Available.ToAdminPB(),
	}, nil
}

func (s *AdminRpcService) SetWalletXagoBalanceEnabled(
	ctx context.Context, request *adminv1.SetWalletXagoBalanceEnabledRequest,
) (*adminv1.Empty, error) {
	walletId := request.GetWalletId()

	w, err := s.b.Wallets().Get(ctx, walletId)
	if err != nil {
		return nil, toGRPCError(err)
	}

	lal, err := s.b.LinkedAccounts().ListByWalletId(ctx, w.ID)
	if err != nil {
		return nil, toGRPCError(err)
	}

	// Check that the linked account doesn't already exist
	for _, la := range lal {
		if la.Provider == xago.ProviderName && la.Type == xago.AccTypeBalance {
			return nil, nil
		}
	}

	await, err := s.b.Xago().CreateBalanceAccount(ctx, xago.CreateBalanceAccArgs{
		WalletID: w.ID,
		Nickname: "ZAR Balance",
		Title:    "ZAR Balance",
		Currency: currency.ZAR,
	})
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
