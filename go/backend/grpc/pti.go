package grpc

import (
	"context"
	"errors"

	"gitlab.com/fynbos/backend/country"
	"gitlab.com/fynbos/backend/linkedaccounts"
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

		ptiBalance, err := s.b.PTI().GetBalance(ctx, la.ID)
		if err != nil {
			return nil, toGRPCError(err)
		}

		resp = append(resp, &pb.PtiBalance{
			Balance:                   ptiBalance.Total.ToPB(),
			Available:                 ptiBalance.Available.ToPB(),
			Currency:                  la.SendCurrency.String(),
			LinkedAccount:             la.ID,
			FormattedBalance:          ptiBalance.Total.Format(),
			FormattedAvailableBalance: ptiBalance.Available.Format(),
		})
	}

	return &backend.GetPtiBalancesResponse{
		Balances: resp,
	}, nil
}

func (s *rpcService) CreatePtiToken(ctx context.Context, req *backend.PtiTokenRequest) (*backend.PtiTokenResponse, error) {
	_, err := s.b.Users().UserForContext(ctx)
	if err != nil && !errors.Is(err, user.ErrNoUserFound) {
		return nil, UnauthenticatedError("Unauthenticated.")
	}

	_, err = s.b.Wallets().ForContext(ctx)
	if err != nil && !errors.Is(err, user.ErrNoUserFound) {
		return nil, ForbiddenError("Unauthenticated.")
	}

	token, err := s.b.PTI().CreateJWT(ctx, pti.TokenArgs{
		URL:    req.GetUrl(),
		Method: req.GetMethod(),
	})
	if err != nil {
		return nil, toGRPCError(err)
	}

	return &backend.PtiTokenResponse{
		AccessToken: token.AccessToken,
		ExpiresAt:   token.ExpiresAt,
		TokenType:   token.TokenType,
	}, nil
}

func (s *rpcService) CreateCard(
	ctx context.Context, req *pb.CreateCardRequest,
) (*pb.LinkedAccount, error) {
	_, err := s.b.Users().UserForContext(ctx)
	if err != nil {
		return nil, UnauthenticatedError("Unauthenticated.")
	}

	w, err := s.b.Wallets().ForContext(ctx)
	if err != nil {
		return nil, UnauthenticatedError("Unauthenticated.")
	}

	feats, err := s.b.Features().Features(ctx, w.ID)
	if err != nil {
		return nil, toGRPCError(err)
	}

	if !feats.AddCardsEnabled {
		return nil, NewValidationError("Form", "You have connected the maximum number of cards to Fynbos.")
	}

	await, err := s.b.PTI().CreateCard(ctx, w.ID, req.GetTokenID())
	if err != nil {
		return nil, toGRPCError(err)
	}

	var la linkedaccounts.LinkedAccount
	err = await(ctx, &la)
	if err != nil {
		return nil, toGRPCError(err)
	}
	if la.ID == "" {
		return nil, toGRPCError(errors.New("linked account not returned from create card workflow"))
	}

	return transformLinkedAccount(la), nil
}

func (s *rpcService) CreatePtiBankAccount(
	ctx context.Context, req *pb.CreatePtiBankAccountRequest,
) (*pb.LinkedAccount, error) {
	_, err := s.b.Users().UserForContext(ctx)
	if err != nil {
		return nil, UnauthenticatedError("Unauthenticated.")
	}

	w, err := s.b.Wallets().ForContext(ctx)
	if err != nil {
		return nil, UnauthenticatedError("Unauthenticated.")
	}

	feats, err := s.b.Features().Features(ctx, w.ID)
	if err != nil {
		return nil, toGRPCError(err)
	}

	if !feats.BanksEnabled {
		return nil, NewValidationError("Form", "You have connected the maximum number of bank accounts to Fynbos.")
	}

	checkDigit := ""
	if len(req.GetRoutingNumber()) > 0 {
		checkDigit = req.GetRoutingNumber()[len(req.GetRoutingNumber())-1:]
	}

	await, err := s.b.PTI().CreateBankAccount(ctx, pti.CreateBankAccountArgs{
		WalletID:                w.ID,
		AccountType:             req.GetAccountType(),
		AccountNumber:           req.GetAccountNumber(),
		RoutingNumber:           req.GetRoutingNumber(),
		RoutingNumberCheckDigit: checkDigit,
		Bank:                    req.GetBankName(),
	})
	if err != nil {
		return nil, toGRPCError(err)
	}

	var la linkedaccounts.LinkedAccount
	err = await(ctx, &la)
	if err != nil {
		return nil, toGRPCError(err)
	}
	if la.ID == "" {
		return nil, toGRPCError(errors.New("linked account not returned from create pti bank account workflow"))
	}

	return transformLinkedAccount(la), nil
}

func (r *rpcService) CreatePTIWithdrawal(ctx context.Context, req *pb.CreatePTIWithdrawalRequest) (*pb.CreatePTIWithdrawalResponse, error) {
	_, err := r.b.Users().UserForContext(ctx)
	if err != nil {
		return nil, UnauthenticatedError("Unauthenticated.")
	}

	wallet, err := r.b.Wallets().ForContext(ctx)
	if err != nil {
		return nil, ForbiddenError("Unauthenticated.")
	}

	if wallet.Country != country.US {
		return nil, toGRPCError(FailedPreconditionError("Wallet not in the US region"))
	}

	p, err := r.b.Payments().Lookup(ctx, req.GetPaymentId())
	if err != nil {
		return nil, toGRPCError(err)
	}

	if p.Sender.WalletID != wallet.ID {
		return nil, NotFoundError("payment not found")
	}
	txID, err := r.b.PTI().CreateWithdrawal(ctx, wallet.ID, p)
	if err != nil {
		return nil, toGRPCError(err)
	}

	return &pb.CreatePTIWithdrawalResponse{
		PaymentId: txID,
	}, nil
}
