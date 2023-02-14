package grpc

import (
	"context"

	"gitlab.com/fynbos/backend/linkedaccounts"

	pb "gitlab.com/fynbos/proto/backend/v1"
)

func (s *rpcService) GetLinkedAccounts(
	ctx context.Context, _ *pb.Empty,
) (*pb.GetLinkedAccountsResponse, error) {
	_, err := s.b.Users().UserForContext(ctx)
	if err != nil {
		return nil, ForbiddenError("Unauthenticated.")
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
		ret[i] = &pb.LinkedAccount{
			Id:   fs.ID,
			Name: fs.Name,
			Mask: fs.Mask,
			Type: fs.Type,
		}
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

	return &pb.LinkedAccount{
		Id:   la.ID,
		Type: la.Type,
		Name: la.Name,
		Mask: la.Mask,
	}, nil

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

	// TODO: One day we will have a switch statement to get the correct provider, but for now it's always machnet
	await, err := s.b.Machnet().DeleteFundSource(ctx, req.Id)
	if err != nil {
		return nil, toGRPCError(err)
	}

	err = await(ctx, nil)

	return &pb.Empty{}, toGRPCError(err)
}

func (s *rpcService) SetNameLinkedAccount(ctx context.Context, req *pb.SetNameLinkedAccountRequest) (*pb.LinkedAccount, error) {
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

	la, err = s.b.LinkedAccounts().SetName(ctx, req.Id, req.Name)
	if err != nil {
		return nil, toGRPCError(err)
	}

	return &pb.LinkedAccount{
		Id:   la.ID,
		Type: la.Type,
		Name: la.Name,
		Mask: la.Mask,
	}, nil
}
