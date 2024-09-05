package grpc

import (
	"context"
	"errors"
	"strings"

	"gitlab.com/fynbos/backend/wallets"
	pb "gitlab.com/fynbos/proto/backend/v1"
)

func (g *rpcService) CreateWalletAddress(ctx context.Context, req *pb.CreateWalletAddressRequest) (*pb.Empty, error) {
	_, err := g.b.Users().UserForContext(ctx)
	if err != nil {
		return nil, UnauthenticatedError("no login found")
	}

	wallet, err := g.b.Wallets().ForContext(ctx)
	if err != nil {
		return nil, ForbiddenError("Unauthenticated.")
	}

	wa, err := wallets.ParseAddress(req.Url)
	if errors.Is(err, wallets.ErrInvalidAddress) {
		return nil, NewValidationError("url", strings.TrimSpace(strings.TrimPrefix(err.Error(), wallets.ErrInvalidAddress.Error())))
	}
	if err != nil {
		return nil, toGRPCError(err)
	}

	_, err = g.b.Wallets().AddAddress(ctx, wallet.ID, wa.String())
	if err != nil {
		return nil, toGRPCError(err)
	}

	wallet, err = g.b.Wallets().SetWalletName(ctx, wallet.ID, req.GetAlias())
	if err != nil {
		return nil, toGRPCError(err)
	}

	err = g.b.Rafiki().CreatePaymentPointer(ctx, *wallet)
	if err != nil {
		return nil, toGRPCError(err)
	}

	return &pb.Empty{}, nil
}

func WalletAddressExists(ctx context.Context, b Backends, addressURLRaw string) (bool, error) {
	// Validate that this is a valid wallet address
	addressURL, err := wallets.ParseAddress(addressURLRaw)
	if err != nil {
		return false, err
	}

	wa, err := b.Wallets().GetFromAddress(ctx, addressURL.String())
	if err != nil && !errors.Is(err, wallets.ErrNoWalletFound) {
		return false, err
	}

	return wa != nil, nil
}

func (g *rpcService) WalletAddressExists(ctx context.Context, req *pb.WalletAddressExistsRequest) (*pb.WalletAddressExistsResponse, error) {
	exists, err := WalletAddressExists(ctx, g.b, req.Url)
	if errors.Is(err, wallets.ErrInvalidAddress) {
		return nil, NewValidationError("url", strings.TrimSpace(strings.TrimPrefix(err.Error(), wallets.ErrInvalidAddress.Error())))
	}
	if err != nil {
		return nil, toGRPCError(err)
	}

	return &pb.WalletAddressExistsResponse{
		Exists: exists,
	}, nil
}
