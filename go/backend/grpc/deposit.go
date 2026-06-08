package grpc

import (
	"context"
	"errors"

	"gitlab.com/fynbos/backend/country"
	"gitlab.com/fynbos/backend/providers/pti"

	"gitlab.com/fynbos/backend/currency"
	"gitlab.com/fynbos/backend/limits"
	"gitlab.com/fynbos/backend/payments"
	"gitlab.com/fynbos/backend/user"
	pb "gitlab.com/fynbos/proto/backend/v1"
)

func (s *rpcService) GetOnOffRampProvider(ctx context.Context, req *pb.Empty) (*pb.GetOnOffRampProviderResponse, error) {
	_, err := s.b.Users().UserForContext(ctx)
	if err != nil && !errors.Is(err, user.ErrNoUserFound) {
		return nil, UnauthenticatedError("Unauthenticated.")
	}

	w, err := s.b.Wallets().ForContext(ctx)
	if err != nil && !errors.Is(err, user.ErrNoUserFound) {
		return nil, ForbiddenError("Unauthenticated.")
	}
	//TODO change to interledger it looks like it's only used for displaying the provider
	provider := "interledger"
	if country.EUCountries[w.Country] {
		provider = "gatehub"
	} else if country.CA == w.Country {
		provider = "chimoney"
	} else if country.US == w.Country {
		provider = pti.ProviderName
	}

	return &pb.GetOnOffRampProviderResponse{
		Provider: provider,
	}, nil
}

func (s *rpcService) DepositBalance(ctx context.Context, req *pb.TransferBalanceRequest) (*pb.Payment, error) {
	_, err := s.b.Users().UserForContext(ctx)
	if err != nil && !errors.Is(err, user.ErrNoUserFound) {
		return nil, UnauthenticatedError("Unauthenticated.")
	}

	w, err := s.b.Wallets().ForContext(ctx)
	if err != nil && !errors.Is(err, user.ErrNoUserFound) {
		return nil, ForbiddenError("Unauthenticated.")
	}

	err = s.validateKYCTransactionRestrictions(ctx, w.ID)
	if err != nil {
		return nil, toGRPCError(err)
	}

	fromLA, err := s.b.LinkedAccounts().Get(ctx, req.FromLinkedAccount)
	if err != nil {
		return nil, toGRPCError(err)
	}
	if fromLA.WalletID != w.ID {
		return nil, NotFoundError("from linked account not found")
	}

	toLA, err := s.b.LinkedAccounts().Get(ctx, req.ToLinkedAccount)
	if err != nil {
		return nil, toGRPCError(err)
	}
	if toLA.WalletID != w.ID {
		return nil, NotFoundError("to linked account not found")
	}

	amt := currency.FromUInt64(req.Amount.Amount, fromLA.SendCurrency)

	// check that does not exceed kyc limits.
	exceedsLimits, limitType, err := s.b.Limits().ExceedsKYCLimits(ctx, w.ID, amt)
	if err != nil {
		return nil, toGRPCError(err)
	}
	if exceedsLimits {
		var description string
		switch limitType {
		case limits.LimitTypeTransaction:
			description = "Exceeds per transaction limit."
		case limits.LimitTypeDaily:
			description = "Exceeds daily limit."
		case limits.LimitTypeMonthly:
			description = "Exceeds monthly limit."
		case limits.LimitType6Monthly:
			description = "Exceeds 6 monthly limit."
		case limits.LimitTypeYearly:
			description = "Exceeds yearly limit."
		default:
			description = "Exceeds account limit."
		}
		return nil, NewValidationError("amount", description)
	}

	p, err := s.b.Payments().Create(ctx, payments.CreateArgs{
		Sender:          payments.Identity{Type: payments.IdentityTypeWalletID, Identifier: w.ID},
		Receiver:        payments.Identity{Type: payments.IdentityTypeWalletID, Identifier: w.ID},
		SenderAmount:    amt,
		SenderAccount:   fromLA.ID,
		ReceiverAmount:  amt,
		ReceiverAccount: toLA.ID,
		Type:            payments.TypeDeposit,
		Note:            req.GetNote(),
	})
	if err != nil {
		return nil, toGRPCError(err)
	}

	return transformPayment(ctx, s.b, p)
}

func (s *rpcService) GetLinkedAccountsForDeposit(ctx context.Context, req *pb.GetLinkedAccountsForTransferRequest) (*pb.GetLinkedAccountsForPaymentResponse, error) {
	_, err := s.b.Users().UserForContext(ctx)
	if err != nil && !errors.Is(err, user.ErrNoUserFound) {
		return nil, UnauthenticatedError("Unauthenticated.")
	}

	w, err := s.b.Wallets().ForContext(ctx)
	if err != nil && !errors.Is(err, user.ErrNoUserFound) {
		return nil, ForbiddenError("Unauthenticated.")
	}

	balance, err := s.b.LinkedAccounts().Get(ctx, req.LinkedAccountId)
	if err != nil {
		return nil, toGRPCError(err)
	}

	lal, err := s.b.LinkedAccounts().ListByWalletId(ctx, w.ID)
	if err != nil {
		return nil, toGRPCError(err)
	}

	var las []*pb.LinkedAccountForPayment
	for _, la := range lal {
		// ListByWalletId returns all rows incl. soft-deleted; skip unlinked accounts.
		if la.DeletedAt.Valid {
			continue
		}
		if balance.Provider == pti.ProviderName && la.Provider == pti.ProviderName && la.Type == pti.TypeBank {
			acc := &pb.LinkedAccountForPayment{
				Details: transformLinkedAccount(la),
				Enabled: la.CanPay(*balance),
			}

			las = append(las, acc)
		}

	}

	return &pb.GetLinkedAccountsForPaymentResponse{
		LinkedAccounts: las,
	}, nil
}

func (s *rpcService) PtiCreateDeposit(ctx context.Context, req *pb.PtiCreateDepositRequest) (*pb.Empty, error) {
	_, err := s.b.Users().UserForContext(ctx)
	if err != nil && !errors.Is(err, user.ErrNoUserFound) {
		return nil, UnauthenticatedError("Unauthenticated.")
	}

	w, err := s.b.Wallets().ForContext(ctx)
	if err != nil && !errors.Is(err, user.ErrNoUserFound) {
		return nil, ForbiddenError("Unauthenticated.")
	}

	p, err := s.b.Payments().Lookup(ctx, req.GetId())
	if err != nil {
		return nil, toGRPCError(err)
	}

	if p.Sender.WalletID != w.ID {
		return nil, NotFoundError("payment not found")
	}

	err = s.b.PTI().CreateDeposit(ctx, w, p)
	if err != nil {
		return nil, toGRPCError(err)
	}
	return &pb.Empty{}, nil
}
