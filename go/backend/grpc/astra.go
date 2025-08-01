package grpc

import (
	"context"
	"errors"
	"time"

	"gitlab.com/fynbos/backend/linkedaccounts"
	"gitlab.com/fynbos/backend/twilio"
	"gitlab.com/fynbos/env"
	"go.temporal.io/sdk/temporal"

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
		return nil, NewValidationError("Form", "You have connected the maximum number of cards to Interledger.")
	}

	las, err := s.b.LinkedAccounts().ListByWalletId(ctx, w.ID)
	if err != nil {
		return nil, toGRPCError(err)
	}
	var linkedCards []linkedaccounts.LinkedAccount
	for _, la := range las {
		if la.Provider == astra.ProviderName && la.Type == astra.TypeCard {
			linkedCards = append(linkedCards, la)
		}
	}

	// limit the number of cards that can be connected to interledger
	// active cards are cards that are not deleted
	// cards created this week are cards that were created in the last week whether they are active or not
	if env.IsProd() {
		var activeCardCount int
		var cardsCreatedWK int
		for _, la := range linkedCards {
			if !la.DeletedAt.Valid {
				activeCardCount++
			}
			if time.Since(la.CreatedAt) < 7*24*time.Hour {
				cardsCreatedWK++
			}
		}
		if activeCardCount >= 5 {
			return nil, CardPreconditionError("cardsVelocityLimit", "You have connected the maximum number of cards to Interledger")
		}
		if cardsCreatedWK >= 2 {
			return nil, CardPreconditionError("cardsMaxLimit", "You have connected the maximum number of cards to Interledger this week")
		}
	}

	await, err := s.b.Astra().CreateCard(ctx, astra.CreateCardArgs{
		WalletID:           w.ID,
		BasisTheoryTokenID: req.GetTokenID(),
	})
	if err != nil {
		return nil, toGRPCError(err)
	}

	var la linkedaccounts.LinkedAccount
	err = await(ctx, &la)
	var applicationError *temporal.ApplicationError
	if errors.As(err, &applicationError) {
		switch applicationError.Type() {
		case "ErrDuplicateCard":
			return nil, AlreadyExistsError("ErrDuplicateCard")
		case "ErrUnsupportedCard":
			return nil, NewValidationError("CardNumber", "Your card is unsupported and cannot be connected to Interledger.")
		case "ErrUnsupportedCountry":
			return nil, NewValidationError("CardNumber", "Your card originates from an unsupported country and cannot be connected to Interledger.")
		case "ErrMultiStatus":
			return nil, UnavailableError("ErrMultiStatus")
		}
	}
	if err != nil {
		return nil, toGRPCError(err)
	}
	if la.ID == "" {
		return nil, toGRPCError(errors.New("linked account not returned from create card workflow"))
	}

	return transformLinkedAccount(la), nil
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
