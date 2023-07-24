package grpc

import (
	"context"
	"fmt"

	"gitlab.com/fynbos/backend/linkedaccounts"
	"gitlab.com/fynbos/backend/providers/tabapay"

	pb "gitlab.com/fynbos/proto/backend/v1"
)

func transformLinkedAccount(la linkedaccounts.LinkedAccount) *pb.LinkedAccount {
	title := la.Nickname
	if la.Nickname == "" {
		title = la.Mask
	}
	return &pb.LinkedAccount{
		Id:         la.ID,
		Type:       la.Type,
		Name:       la.Name,
		Mask:       la.Mask,
		Nickname:   la.Nickname,
		CanSend:    la.CanSend,
		CanReceive: la.CanReceive,
		Title:      title,
	}
}
func (s *rpcService) GetLinkedAccounts(
	ctx context.Context, _ *pb.Empty,
) (*pb.GetLinkedAccountsResponse, error) {
	_, err := s.b.Users().UserForContext(ctx)
	if err != nil {
		return nil, UnauthenticatedError("Unauthenticated.")
	}

	wallet, err := s.b.Users().WalletForContext(ctx)
	if err != nil {
		return nil, ForbiddenError("Unauthenticated.")
	}

	linkedAccounts, err := s.b.LinkedAccounts().ListByWalletId(ctx, wallet.ID)
	if err != nil {
		return nil, InternalError("Unable to get linked accounts.")
	}

	ret := make([]*pb.LinkedAccount, len(linkedAccounts))
	for i, fs := range linkedAccounts {
		ret[i] = transformLinkedAccount(fs)
	}

	return &pb.GetLinkedAccountsResponse{
		LinkedAccounts: ret,
	}, nil
}

func (s *rpcService) GetLinkedAccount(ctx context.Context, req *pb.GetLinkedAccountRequest) (*pb.LinkedAccount, error) {
	_, err := s.b.Users().UserForContext(ctx)
	if err != nil {
		return nil, UnauthenticatedError("Unauthenticated.")
	}

	wallet, err := s.b.Users().WalletForContext(ctx)
	if err != nil {
		return nil, ForbiddenError("Unauthenticated.")
	}

	la, err := s.b.LinkedAccounts().Get(ctx, req.Id)
	if err != nil {
		return nil, toGRPCError(err)
	}

	if la.WalletID != wallet.ID {
		return nil, toGRPCError(linkedaccounts.ErrNotFound)
	}

	return transformLinkedAccount(*la), nil

}

func (s *rpcService) DeleteLinkedAccount(ctx context.Context, req *pb.DeleteLinkedAccountRequest) (*pb.Empty, error) {
	_, err := s.b.Users().UserForContext(ctx)
	if err != nil {
		return nil, UnauthenticatedError("Unauthenticated.")
	}

	wallet, err := s.b.Users().WalletForContext(ctx)
	if err != nil {
		return nil, ForbiddenError("Unauthenticated.")
	}

	la, err := s.b.LinkedAccounts().Get(ctx, req.Id)
	if err != nil {
		return nil, toGRPCError(err)
	}

	if la.WalletID != wallet.ID {
		return nil, toGRPCError(linkedaccounts.ErrNotFound)
	}

	// TODO: Add provider for the linked account here.

	return &pb.Empty{}, toGRPCError(err)
}

func (s *rpcService) SetNicknameLinkedAccount(ctx context.Context, req *pb.SetNicknameLinkedAccountRequest) (*pb.LinkedAccount, error) {
	_, err := s.b.Users().UserForContext(ctx)
	if err != nil {
		return nil, UnauthenticatedError("Unauthenticated.")
	}

	wallet, err := s.b.Users().WalletForContext(ctx)
	if err != nil {
		return nil, ForbiddenError("Unauthenticated.")
	}

	// ACL check if user and wallet owns linked account
	la, err := s.b.LinkedAccounts().Get(ctx, req.Id)
	if err != nil {
		return nil, toGRPCError(err)
	}
	if la.WalletID != wallet.ID {
		return nil, toGRPCError(linkedaccounts.ErrNotFound)
	}

	la, err = s.b.LinkedAccounts().SetNickname(ctx, req.Id, req.Nickname)
	if err != nil {
		return nil, toGRPCError(err)
	}

	return transformLinkedAccount(*la), nil
}

func (s *rpcService) GetCardDetails(ctx context.Context, req *pb.GetCardDetailsRequest) (*pb.CardDetails, error) {
	_, err := s.b.Users().UserForContext(ctx)
	if err != nil {
		return nil, UnauthenticatedError("Unauthenticated.")
	}

	wallet, err := s.b.Users().WalletForContext(ctx)
	if err != nil {
		return nil, ForbiddenError("Unauthenticated.")
	}

	la, err := s.b.LinkedAccounts().Get(ctx, req.GetId())
	if err != nil {
		return nil, toGRPCError(err)
	}
	if la.WalletID != wallet.ID && la.Type != tabapay.ProviderName {
		return nil, NotFoundError("ErrNotFound")
	}

	card, err := s.b.BasisTheory().GetCard(ctx, req.GetId())
	if err != nil {
		return nil, toGRPCError(err)
	}
	if card.WalletID != wallet.ID {
		return nil, NotFoundError("ErrNotFound")
	}

	network := card.PullNetwork
	if network == "" {
		network = card.PushNetwork
	}
	cardType := card.PullType
	if cardType == "" {
		cardType = card.PushType
	}

	return &pb.CardDetails{
		Id:         card.ID,
		Network:    network,
		Bin:        card.Bin,
		Type:       cardType,
		Expiration: fmt.Sprintf("%s/%s", card.ExpirationMonth, card.ExpirationYear),
		Last4:      la.Mask,
		Nickname:   la.Nickname,
		State:      string(la.State),
		CanSend:    la.CanSend,
		CanReceive: la.CanReceive,
	}, nil
}
