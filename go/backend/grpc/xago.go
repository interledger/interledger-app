package grpc

import (
	"context"
	"errors"

	"gitlab.com/fynbos/backend/currency"
	"gitlab.com/fynbos/backend/limits"
	"gitlab.com/fynbos/backend/linkedaccounts"
	"gitlab.com/fynbos/backend/payments"
	"gitlab.com/fynbos/backend/providers/xago"
	"gitlab.com/fynbos/backend/user"
	"gitlab.com/fynbos/env"
	"gitlab.com/fynbos/log"
	pb "gitlab.com/fynbos/proto/backend/v1"
	"go.uber.org/zap"
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

	await, err := s.b.Xago().CreateBalanceAccount(ctx, xago.CreateBalanceAccArgs{
		WalletID: w.ID,
		Nickname: "ZAR Balance",
		Title:    "ZAR Balance",
		Currency: cc,
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

func (s *rpcService) WithdrawXagoBalance(ctx context.Context, req *pb.WithdrawXagoBalanceRequest) (*pb.Payment, error) {
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
	if fromLA.WalletID != w.ID || fromLA.Provider != xago.ProviderName || fromLA.Type != xago.AccTypeBalance {
		return nil, NotFoundError("from linked account not found for xago")
	}

	toLA, err := s.b.LinkedAccounts().Get(ctx, req.ToLinkedAccount)
	if err != nil {
		return nil, toGRPCError(err)
	}
	if toLA.WalletID != w.ID || toLA.Provider != xago.ProviderName || toLA.Type != xago.AccTypeBank {
		return nil, NotFoundError("to linked account not found for xago")
	}

	amt := currency.FromPB(req.Amount)

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

func (s *rpcService) GetXagoBalances(ctx context.Context, req *pb.Empty) (*pb.GetXagoBalanceResponse, error) {
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

	var resp []*pb.XagoBalance
	for _, la := range lal {
		if la.Provider != xago.ProviderName || la.Type != xago.AccTypeBalance {
			continue
		}

		bal, err := s.b.Xago().GetBalance(ctx, la.ID)
		if err != nil {
			return nil, toGRPCError(err)
		}

		resp = append(resp, &pb.XagoBalance{
			Balance:                   bal.Total.ToPB(),
			Available:                 bal.Available.ToPB(),
			Currency:                  la.SendCurrency.String(),
			LinkedAccount:             la.ID,
			FormattedBalance:          bal.Total.Format(),
			FormattedAvailableBalance: bal.Available.Format(),
		})
	}

	return &pb.GetXagoBalanceResponse{Balances: resp}, nil
}

func (s *rpcService) GetXagoDepositDetails(ctx context.Context, req *pb.GetXagoDepositDetailsRequest) (*pb.GetXagoDepositDetailsResponse, error) {
	_, err := s.b.Users().UserForContext(ctx)
	if err != nil && !errors.Is(err, user.ErrNoUserFound) {
		return nil, UnauthenticatedError("Unauthenticated.")
	}

	w, err := s.b.Wallets().ForContext(ctx)
	if err != nil && !errors.Is(err, user.ErrNoUserFound) {
		return nil, ForbiddenError("Unauthenticated.")
	}

	la, err := s.b.LinkedAccounts().Get(ctx, req.LinkedAccount)
	if err != nil {
		return nil, toGRPCError(err)
	}

	if la.WalletID != w.ID || la.Provider != xago.ProviderName || la.Type != xago.AccTypeBalance {
		return nil, NotFoundError("linked account not found")
	}

	sa, err := s.b.Xago().LookupSubAccount(ctx, w.ID)
	if err != nil {
		return nil, toGRPCError(err)
	}

	var resp []*pb.XagoDepositDetails
	account, err := s.b.Xago().GetBankAccount(ctx)
	if err != nil {
		return nil, toGRPCError(err)
	}
	resp = append(resp, &pb.XagoDepositDetails{
		Currency:         account.CurrencyCode.String(),
		AccountNumber:    account.AccountNumber,
		BranchCode:       account.BranchCode,
		BankName:         account.BankName,
		DepositReference: sa.DepositReference,
	})

	return &pb.GetXagoDepositDetailsResponse{Details: resp}, nil
}

func (s *rpcService) DepositTestXago(ctx context.Context, req *pb.Empty) (*pb.Empty, error) {
	u, err := s.b.Users().UserForContext(ctx)
	if err != nil && !errors.Is(err, user.ErrNoUserFound) {
		return nil, UnauthenticatedError("Unauthenticated.")
	}

	if !env.FeatureTestingTestDeposits() {
		log.Warn("received xago test deposit RPC call in non-testing environment", zap.String("userId", u.ID))
		return nil, ForbiddenError("Forbidden.")
	}

	w, err := s.b.Wallets().ForContext(ctx)
	if err != nil && !errors.Is(err, user.ErrNoUserFound) {
		return nil, ForbiddenError("Unauthenticated.")
	}

	sa, err := s.b.Xago().LookupSubAccount(ctx, w.ID)
	if err != nil {
		return nil, toGRPCError(err)
	}

	err = s.b.Xago().TestDeposit(ctx, *sa)
	if err != nil {
		return nil, toGRPCError(err)
	}

	return &pb.Empty{}, nil
}
