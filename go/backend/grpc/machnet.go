package grpc

import (
	"context"
	"errors"

	"gitlab.com/fynbos/backend/linkedaccounts"
	"gitlab.com/fynbos/backend/providers/machnet"
	backendv1 "gitlab.com/fynbos/proto/backend/v1"
)

func (s *rpcService) GetMachnetWidgetToken(
	ctx context.Context, req *backendv1.Empty,
) (*backendv1.MachnetWidgetToken, error) {
	_, err := s.b.Users().UserForContext(ctx)
	if err != nil {
		return nil, ForbiddenError("Unauthenticated.")
	}

	wallet, err := s.b.Users().WalletForContext(ctx)
	if err != nil {
		return nil, ForbiddenError("Unauthenticated.")
	}

	token, err := s.b.Machnet().GetWidgetToken(ctx, wallet.ID)
	if err != nil {
		return nil, toGRPCError(err)
	}

	return &backendv1.MachnetWidgetToken{
		Value:            token.Value,
		ExpiresInMinutes: int64(token.ExpiresInMinutes),
		UserId:           token.UserID,
	}, nil
}

func (s *rpcService) CreateSendUser(
	ctx context.Context, req *backendv1.Empty,
) (*backendv1.Empty, error) {
	_, err := s.b.Users().UserForContext(ctx)
	if err != nil {
		return nil, ForbiddenError("Unauthenticated.")
	}

	wallet, err := s.b.Users().WalletForContext(ctx)
	if err != nil {
		return nil, ForbiddenError("Unauthenticated.")
	}

	await, err := s.b.Machnet().CreateSendUser(ctx, wallet.ID)
	if err != nil {
		return nil, toGRPCError(err)
	}

	err = await(ctx, nil)
	if err != nil {
		return nil, toGRPCError(err)
	}

	return &backendv1.Empty{}, nil
}

func (s *rpcService) HasSendUser(
	ctx context.Context, req *backendv1.Empty,
) (*backendv1.HasSendUserResponse, error) {
	_, err := s.b.Users().UserForContext(ctx)
	if err != nil {
		return nil, ForbiddenError("Unauthenticated.")
	}

	wallet, err := s.b.Users().WalletForContext(ctx)
	if err != nil {
		return nil, ForbiddenError("Unauthenticated.")
	}

	_, err = s.b.Machnet().GetUserByWalletID(ctx, wallet.ID)
	if err != nil {
		if errors.Is(err, machnet.ErrNotFound) {
			return &backendv1.HasSendUserResponse{
				HasSendUser: false,
			}, nil
		}
		return nil, toGRPCError(err)
	}

	return &backendv1.HasSendUserResponse{
		HasSendUser: true,
	}, nil
}

func (s *rpcService) KYCStatus(ctx context.Context, _ *backendv1.Empty) (*backendv1.KYCStatusResponse, error) {
	_, err := s.b.Users().UserForContext(ctx)
	if err != nil {
		return nil, UnauthenticatedError("Unauthenticated.")
	}

	wallet, err := s.b.Users().WalletForContext(ctx)
	if err != nil {
		return nil, ForbiddenError("Unauthorized.")
	}

	kyc, err := s.b.Machnet().GetKYCStatus(ctx, wallet.ID)
	if err != nil {
		if errors.Is(err, machnet.ErrNotFound) {
			return &backendv1.KYCStatusResponse{
				HasSendUser: false,
			}, nil
		}
		return nil, toGRPCError(err)
	}

	return &backendv1.KYCStatusResponse{
		HasSendUser:  true,
		KycStatus:    int32(kyc.User.KYCStatus),
		FailedFields: kyc.FailedFields,
	}, nil
}

func (s *rpcService) CreateWallet(
	ctx context.Context, req *backendv1.CreateWalletRequest,
) (*backendv1.LinkedAccount, error) {
	_, err := s.b.Users().UserForContext(ctx)
	if err != nil {
		return nil, UnauthenticatedError("Unauthenticated.")
	}

	wallet, err := s.b.Users().WalletForContext(ctx)
	if err != nil {
		return nil, UnauthenticatedError("Unauthenticated.")
	}

	if err = s.b.Validator().VarCtx(ctx, req.GetNickname(), "required"); err != nil {
		return nil, toGRPCError(err)
	}

	sendUser, err := s.b.Machnet().GetUserByWalletID(ctx, wallet.ID)
	if err != nil {
		return nil, toGRPCError(err)
	}

	la, err := s.b.Machnet().CreateWallet(ctx, machnet.CreateWalletArgs{
		Nickname:   req.GetNickname(),
		SendUserID: sendUser.ID,
	})
	if err != nil {
		return nil, toGRPCError(err)
	}

	return &backendv1.LinkedAccount{
		Id:   la.ID,
		Type: la.Type,
		Name: la.Name,
		Mask: la.Mask,
	}, nil
}

func (s *rpcService) GetWalletBalance(ctx context.Context, _ *backendv1.Empty) (*backendv1.WalletBalance, error) {
	_, err := s.b.Users().UserForContext(ctx)
	if err != nil {
		return nil, UnauthenticatedError("Unauthenticated.")
	}

	wallet, err := s.b.Users().WalletForContext(ctx)
	if err != nil {
		return nil, ForbiddenError("Unauthenticated.")
	}

	lal, err := s.b.LinkedAccounts().ListByWalletId(ctx, wallet.ID)
	if err != nil {
		return nil, toGRPCError(err)
	}

	var found *linkedaccounts.LinkedAccount
	for _, la := range lal {
		if la.Provider != machnet.ProviderName || la.Type != machnet.TypeWallet {
			continue
		}
		found = &la
		break
	}
	if found == nil {
		return nil, NotFoundError("machnet wallet not found")
	}

	mw, err := s.b.Machnet().GetWallet(ctx, found.ProviderID)
	if err != nil {
		return nil, toGRPCError(err)
	}

	return &backendv1.WalletBalance{
		Balance:   mw.Balance,
		Available: mw.AvailableBalance,
	}, nil
}

type validateWithdrawFromMachnetWalletArgs struct {
	ToLinkedAccount string `validate:"required,uuid"`
	Amount          uint64 `validate:"gt=0"`
	IpAddress       string `validate:"ip_addr"`
}

func (s *rpcService) WithdrawFromMachnetWallet(
	ctx context.Context, req *backendv1.WithdrawFromMachnetWalletRequest,
) (*backendv1.MachnetWalletWithdrawal, error) {
	_, err := s.b.Users().UserForContext(ctx)
	if err != nil {
		return nil, UnauthenticatedError("Unauthenticated.")
	}

	wallet, err := s.b.Users().WalletForContext(ctx)
	if err != nil {
		return nil, UnauthenticatedError("Unauthenticated.")
	}

	if err = s.b.Validator().StructCtx(ctx, validateWithdrawFromMachnetWalletArgs{
		ToLinkedAccount: req.GetToLinkedAccountId(),
		Amount:          req.GetAmount(),
		IpAddress:       req.GetIpAddress(),
	}); err != nil {
		return nil, toGRPCError(err)
	}

	toLinkedAcc, err := s.b.LinkedAccounts().Get(ctx, req.GetToLinkedAccountId())
	if err != nil {
		return nil, toGRPCError(err)
	}
	if toLinkedAcc.WalletID != wallet.ID {
		return nil, NotFoundError("Linked account not found.")
	}

	linkedAccounts, err := s.b.LinkedAccounts().ListByWalletId(ctx, wallet.ID)
	if err != nil {
		return nil, toGRPCError(err)
	}
	var linkedWallet *linkedaccounts.LinkedAccount
	for _, la := range linkedAccounts {
		if la.Provider == machnet.ProviderName && la.Type == machnet.TypeWallet {
			linkedWallet = &la
			break
		}
	}
	if linkedWallet == nil {
		return nil, toGRPCError(errors.New("Machnet wallet not found."))
	}

	withdrawal, err := s.b.Machnet().WithdrawFromWallet(ctx, machnet.WithdrawFromWalletArgs{
		Amount:                req.GetAmount(),
		WalletLinkedAccountID: linkedWallet.ID,
		ToLinkedAccountID:     req.GetToLinkedAccountId(),
		IpAddress:             req.GetIpAddress(),
	})
	if err != nil {
		return nil, toGRPCError(err)
	}

	return &backendv1.MachnetWalletWithdrawal{
		Id:                withdrawal.ID,
		Amount:            withdrawal.Amount,
		ToLinkedAccountId: toLinkedAcc.ID,
		Status:            withdrawal.Status,
	}, nil
}
