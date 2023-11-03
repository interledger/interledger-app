package grpc

import (
	"context"
	"errors"
	"fmt"

	"gitlab.com/fynbos/backend/country"
	"gitlab.com/fynbos/backend/currency"

	"gitlab.com/fynbos/backend/linkedaccounts"

	"gitlab.com/fynbos/backend/providers/xago"

	"gitlab.com/fynbos/backend/user"

	pb "gitlab.com/fynbos/proto/backend/v1"
)

func (s *rpcService) AddXagoBankAccount(ctx context.Context, req *pb.AddXagoBankAccountRequest) (*pb.LinkedAccount, error) {
	_, err := s.b.Users().UserForContext(ctx)
	if err != nil && !errors.Is(err, user.ErrNoUserFound) {
		return nil, UnauthenticatedError("Unauthenticated.")
	}

	w, err := s.b.Wallets().ForContext(ctx)
	if err != nil && !errors.Is(err, user.ErrNoUserFound) {
		return nil, ForbiddenError("Unauthenticated.")
	}

	await, err := s.b.Xago().CreateBeneficiary(ctx, xago.CreateBankAccountArgs{
		WalletID:      w.ID,
		AccountNumber: req.AccountNumber,
		BranchCode:    req.BranchCode,
		BankName:      req.BankName,
		IBAN:          req.Iban,
		BIC:           req.Bic,
	})
	if err != nil {
		return nil, toGRPCError(err)
	}

	var la linkedaccounts.LinkedAccount
	err = await(ctx, &la)
	if err != nil {
		return nil, toGRPCError(err)
	}

	return transformLinkedAccount(la), nil
}

func (s *rpcService) AddXagoBalanceAccount(ctx context.Context, req *pb.AddXagoBalanceAccountRequest) (*pb.LinkedAccount, error) {
	_, err := s.b.Users().UserForContext(ctx)
	if err != nil && !errors.Is(err, user.ErrNoUserFound) {
		return nil, UnauthenticatedError("Unauthenticated.")
	}

	w, err := s.b.Wallets().ForContext(ctx)
	if err != nil && !errors.Is(err, user.ErrNoUserFound) {
		return nil, ForbiddenError("Unauthenticated.")
	}

	if req.CurrencyCode != currency.ZAR.String() && req.CurrencyCode != currency.USD.String() {
		return nil, FailedPreconditionError("unsupported currency")
	}

	lal, err := s.b.LinkedAccounts().ListByWalletId(ctx, w.ID)
	if err != nil {
		return nil, toGRPCError(err)
	}

	// Check that the linked account doesn't already exist
	for _, la := range lal {
		if la.Provider == xago.ProviderName && la.Type == xago.AccTypeBalance && la.SendCurrency.String() == req.CurrencyCode {
			return transformLinkedAccount(la), nil
		}
	}

	cc := currency.ParseCurrency(req.CurrencyCode)
	nation := country.ZA
	if cc == currency.USD {
		nation = country.US
	}

	la, err := s.b.LinkedAccounts().Create(ctx, &linkedaccounts.CreateArgs{
		WalletID:        w.ID,
		Name:            req.Title,
		Nickname:        req.Nickname,
		Provider:        xago.ProviderName,
		ProviderID:      fmt.Sprintf("xago_%s_%s", cc.String(), w.ID), // Deterministic providerID stops duplicate accounts from being created.
		Type:            xago.AccTypeBalance,
		CanSend:         true,
		CanReceive:      true,
		State:           linkedaccounts.Verified,
		SendCountry:     nation,
		SendCurrency:    cc,
		ReceiveCountry:  nation,
		ReceiveCurrency: cc,
	})
	if err != nil {
		return nil, toGRPCError(err)
	}

	return transformLinkedAccount(*la), nil
}
