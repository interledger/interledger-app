package grpc

import (
	"context"
	"errors"
	"time"

	"gitlab.com/fynbos/backend/twilio"

	"gitlab.com/fynbos/backend/currency"
	"gitlab.com/fynbos/backend/payments"
	"gitlab.com/fynbos/backend/providers/astra"
	"gitlab.com/fynbos/backend/providers/pti"
	"gitlab.com/fynbos/backend/user"
	pb "gitlab.com/fynbos/proto/backend/v1"
)

func (s *rpcService) AstraRequiresOTP(
	ctx context.Context, req *pb.Empty,
) (*pb.AstraRequiresOTPResponse, error) {
	_, err := s.b.Users().UserForContext(ctx)
	if err != nil {
		return nil, UnauthenticatedError("Unauthenticated.")
	}

	w, err := s.b.Wallets().ForContext(ctx)
	if err != nil {
		return nil, UnauthenticatedError("Unauthenticated.")
	}

	// OTP is required to have been sent in the past 30 days. We'll refresh pre-emptively and use
	// 29 days.
	usrs, err := s.b.Users().ListUsers(ctx, w.ID)
	if err != nil {
		return nil, toGRPCError(err)
	}
	if len(usrs) < 1 {
		return nil, InternalError("no user found for wallet")
	}

	verifications, err := s.b.Twilio().ListSuccessfulVerificationAttempts(ctx, twilio.ListSuccessfulVerificationAttemptsArgs{
		To:    usrs[0].PhoneNumber,
		Limit: 1,
		After: time.Now().AddDate(0, 0, -29),
	})
	if err != nil {
		return nil, toGRPCError(err)
	}

	return &pb.AstraRequiresOTPResponse{IsRequired: len(verifications) > 1}, nil
}

func (s *rpcService) AstraDepositFromCard(ctx context.Context, req *pb.AstraDepositFromCardRequest) (*pb.Payment, error) {
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
	if fromLA.WalletID != w.ID || fromLA.Provider != astra.ProviderName || fromLA.Type != astra.TypeCard {
		return nil, NotFoundError("from linked account not found for astra")
	}

	toLA, err := s.b.LinkedAccounts().Get(ctx, req.ToLinkedAccount)
	if err != nil {
		return nil, toGRPCError(err)
	}
	if toLA.WalletID != w.ID || toLA.Provider != pti.ProviderName || toLA.Type != pti.AccTypeBalance {
		return nil, NotFoundError("to linked account not found for PTI")
	}

	p, err := s.b.Payments().Create(ctx, payments.CreateArgs{
		Sender:          payments.Identity{Type: payments.IdentityTypeWalletID, Identifier: w.ID},
		Receiver:        payments.Identity{Type: payments.IdentityTypeWalletID, Identifier: w.ID},
		SenderAmount:    currency.FromPB(req.Amount),
		SenderAccount:   fromLA.ID,
		ReceiverAmount:  currency.FromPB(req.Amount),
		ReceiverAccount: toLA.ID,
		Type:            payments.TypeDeposit,
		Note:            req.GetNote(),
	})
	if err != nil {
		return nil, toGRPCError(err)
	}

	return transformPayment(ctx, s.b, p)
}

func (s *rpcService) AstraWithdrawToCard(ctx context.Context, req *pb.AstraWithdrawToCardRequest) (*pb.Payment, error) {
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
	if fromLA.WalletID != w.ID || fromLA.Provider != pti.ProviderName || fromLA.Type != pti.AccTypeBalance {
		return nil, NotFoundError("from linked account not found for pti")
	}

	toLA, err := s.b.LinkedAccounts().Get(ctx, req.ToLinkedAccount)
	if err != nil {
		return nil, toGRPCError(err)
	}
	if toLA.WalletID != w.ID || toLA.Provider != astra.ProviderName || toLA.Type != astra.TypeCard {
		return nil, NotFoundError("to linked account not found for astr")
	}

	p, err := s.b.Payments().Create(ctx, payments.CreateArgs{
		Sender:          payments.Identity{Type: payments.IdentityTypeWalletID, Identifier: w.ID},
		Receiver:        payments.Identity{Type: payments.IdentityTypeWalletID, Identifier: w.ID},
		SenderAmount:    currency.FromPB(req.Amount),
		SenderAccount:   fromLA.ID,
		ReceiverAmount:  currency.FromPB(req.Amount),
		ReceiverAccount: toLA.ID,
		Type:            payments.TypeWithdrawal,
		Note:            req.GetNote(),
	})
	if err != nil {
		return nil, toGRPCError(err)
	}

	return transformPayment(ctx, s.b, p)
}
