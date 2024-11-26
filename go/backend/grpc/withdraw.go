package grpc

import (
	"context"
	"errors"

	"gitlab.com/fynbos/backend/providers/chimoney"
	"gitlab.com/fynbos/backend/providers/pti"

	"gitlab.com/fynbos/backend/currency"
	"gitlab.com/fynbos/backend/limits"
	"gitlab.com/fynbos/backend/payments"
	"gitlab.com/fynbos/backend/providers/xago"
	"gitlab.com/fynbos/backend/user"
	pb "gitlab.com/fynbos/proto/backend/v1"
)

func (s *rpcService) WithdrawBalance(ctx context.Context, req *pb.TransferBalanceRequest) (*pb.Payment, error) {
	_, err := s.b.Users().UserForContext(ctx)
	if err != nil && !errors.Is(err, user.ErrNoUserFound) {
		return nil, UnauthenticatedError("Unauthenticated.")
	}

	w, err := s.b.Wallets().ForContext(ctx)
	if err != nil && !errors.Is(err, user.ErrNoUserFound) {
		return nil, ForbiddenError("Unauthenticated.")
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
	exceedsLimits, limitType, err := s.b.Limits().ExceedsKYCLimits(ctx, w.ID, currency.FromUInt64(0, amt.Currency))
	if err != nil {
		return nil, toGRPCError(err)
	}
	if exceedsLimits {
		var description string
		switch limitType {
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
		Type:            payments.TypeWithdrawal,
		Note:            req.GetNote(),
	})
	if err != nil {
		return nil, toGRPCError(err)
	}

	return transformPayment(ctx, s.b, p)
}

func (s *rpcService) GetLinkedAccountsForWithdraw(ctx context.Context, req *pb.GetLinkedAccountsForTransferRequest) (*pb.GetLinkedAccountsForPaymentResponse, error) {
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
		if balance.Provider == xago.ProviderName && la.Provider == xago.ProviderName && la.Type == xago.AccTypeBank {
			acc := &pb.LinkedAccountForPayment{
				Details: transformLinkedAccount(la),
				Enabled: balance.CanPay(la),
			}

			las = append(las, acc)
		}

		if balance.Provider == pti.ProviderName && la.Provider == pti.ProviderName && la.Type == pti.TypeCard {
			acc := &pb.LinkedAccountForPayment{
				Details: transformLinkedAccount(la),
				Enabled: balance.CanPay(la),
			}

			las = append(las, acc)
		}

		if balance.Provider == chimoney.ProviderName && la.Provider == chimoney.ProviderName && la.Type == chimoney.AccTypeInterac {
			acc := &pb.LinkedAccountForPayment{
				Details: transformLinkedAccount(la),
				Enabled: balance.CanPay(la),
			}

			las = append(las, acc)
		}

	}

	return &pb.GetLinkedAccountsForPaymentResponse{
		LinkedAccounts: las,
	}, nil
}
