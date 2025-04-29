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

func WalletAddressExists(ctx context.Context, b Backends, addressURL string) (bool, error) {
	wa, err := b.Wallets().GetFromAddress(ctx, addressURL)
	if err != nil && !errors.Is(err, wallets.ErrNoWalletFound) {
		return false, err
	}

	return wa != nil, nil
}

func (g *rpcService) WalletAddressValidAndNotExists(ctx context.Context, req *pb.WalletAddressExistsRequest) (*pb.WalletAddressExistsResponse, error) {
	addressURLRaw := req.Url
	// Validate that the wallet address matches validation criteria
	addressURL, err := wallets.ParseAddress(addressURLRaw)
	if err != nil {
		return nil, NewValidationError("url", strings.TrimSpace(strings.TrimPrefix(err.Error(), wallets.ErrInvalidAddress.Error())))
	}

	exists, err := WalletAddressExists(ctx, g.b, addressURL.String())
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
