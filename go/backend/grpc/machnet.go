package grpc

import (
	"context"
	"errors"

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
