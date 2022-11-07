package grpc

import (
	"context"
	"errors"

	"gitlab.com/fynbos/backend/linkedaccounts"
	"gitlab.com/fynbos/backend/providers/machnet"
	backendv1 "gitlab.com/fynbos/proto/backend/v1"
)

func (r *rpcService) GetMachnetWidgetToken(
	ctx context.Context, req *backendv1.Empty,
) (*backendv1.MachnetWidgetToken, error) {
	_, err := r.b.Users().UserForContext(ctx)
	if err != nil {
		return nil, ForbiddenError("Unauthenticated.")
	}

	wallet, err := r.b.Users().WalletForContext(ctx)
	if err != nil {
		return nil, ForbiddenError("Unauthenticated.")
	}

	token, err := r.b.Machnet().GetWidgetToken(ctx, wallet.ID)
	if err != nil {
		return nil, toGRPCError(err)
	}

	return &backendv1.MachnetWidgetToken{
		Value:            token.Value,
		ExpiresInMinutes: int64(token.ExpiresInMinutes),
		UserId:           token.UserID,
	}, nil
}

func (r *rpcService) CreateSendUser(
	ctx context.Context, req *backendv1.Empty,
) (*backendv1.Empty, error) {
	_, err := r.b.Users().UserForContext(ctx)
	if err != nil {
		return nil, ForbiddenError("Unauthenticated.")
	}

	wallet, err := r.b.Users().WalletForContext(ctx)
	if err != nil {
		return nil, ForbiddenError("Unauthenticated.")
	}

	await, err := r.b.Machnet().CreateSendUser(ctx, wallet.ID)
	if err != nil {
		return nil, toGRPCError(err)
	}

	err = await(ctx, nil)
	if err != nil {
		return nil, toGRPCError(err)
	}

	return &backendv1.Empty{}, nil
}

func (r *rpcService) HasSendUser(
	ctx context.Context, req *backendv1.Empty,
) (*backendv1.HasSendUserResponse, error) {
	_, err := r.b.Users().UserForContext(ctx)
	if err != nil {
		return nil, ForbiddenError("Unauthenticated.")
	}

	wallet, err := r.b.Users().WalletForContext(ctx)
	if err != nil {
		return nil, ForbiddenError("Unauthenticated.")
	}

	_, err = r.b.Machnet().GetUserByWalletID(ctx, wallet.ID)
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

func (r *rpcService) CreateWallet(
	ctx context.Context, req *backendv1.CreateWalletRequest,
) (*backendv1.LinkedAccount, error) {
	_, err := r.b.Users().UserForContext(ctx)
	if err != nil {
		return nil, UnauthenticatedError("Unauthenticated.")
	}

	wallet, err := r.b.Users().WalletForContext(ctx)
	if err != nil {
		return nil, UnauthenticatedError("Unauthenticated.")
	}

	sendUser, err := r.b.Machnet().GetUserByWalletID(ctx, wallet.ID)
	if err != nil {
		return nil, toGRPCError(err)
	}

	la, err := r.b.Machnet().CreateWallet(ctx, machnet.CreateWalletArgs{
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

func (r *rpcService) GetWalletBalance(ctx context.Context, _ *backendv1.Empty) (*backendv1.WalletBalance, error) {
	_, err := r.b.Users().UserForContext(ctx)
	if err != nil {
		return nil, UnauthenticatedError("Unauthenticated.")
	}

	wallet, err := r.b.Users().WalletForContext(ctx)
	if err != nil {
		return nil, ForbiddenError("Unauthenticated.")
	}

	lal, err := r.b.LinkedAccounts().ListByWalletId(ctx, wallet.ID)
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

	mw, err := r.b.Machnet().GetWallet(ctx, found.ProviderID)
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

func (r *rpcService) WithdrawFromMachnetWallet(
	ctx context.Context, req *backendv1.WithdrawFromMachnetWalletRequest,
) (*backendv1.MachnetWalletWithdrawal, error) {
	_, err := r.b.Users().UserForContext(ctx)
	if err != nil {
		return nil, UnauthenticatedError("Unauthenticated.")
	}

	wallet, err := r.b.Users().WalletForContext(ctx)
	if err != nil {
		return nil, UnauthenticatedError("Unauthenticated.")
	}

	if err = r.b.Validator().Struct(validateWithdrawFromMachnetWalletArgs{
		ToLinkedAccount: req.GetToLinkedAccountId(),
		Amount:          req.GetAmount(),
		IpAddress:       req.GetIpAddress(),
	}); err != nil {
		return nil, ValidationError(err, validationDesc)
	}

	toLinkedAcc, err := r.b.LinkedAccounts().Get(ctx, req.GetToLinkedAccountId())
	if err != nil {
		return nil, toGRPCError(err)
	}
	if toLinkedAcc.WalletID != wallet.ID {
		return nil, NotFoundError("Linked account not found.")
	}

	linkedAccounts, err := r.b.LinkedAccounts().ListByWalletId(ctx, wallet.ID)
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

	withdrawal, err := r.b.Machnet().WithdrawFromWallet(ctx, machnet.WithdrawFromWalletArgs{
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
