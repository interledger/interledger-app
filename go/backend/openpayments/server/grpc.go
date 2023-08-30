package server

import (
	"context"
	"errors"
	"strings"

	"gitlab.com/fynbos/backend/openpayments/ops"
	"gitlab.com/fynbos/backend/wallets"
	pb "gitlab.com/fynbos/proto/backend/v1"
)

var _ pb.OpenPaymentServiceServer = &grpcServer{}

type grpcServer struct {
	b Backends
}

func NewGRPCServer(b Backends) pb.OpenPaymentServiceServer {
	return &grpcServer{b: b}
}

func (g *grpcServer) CreatePaymentPointer(ctx context.Context, req *pb.CreatePaymentPointerRequest) (*pb.Empty, error) {
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

	_, err = g.b.Wallets().SetWalletName(ctx, wallet.ID, req.GetAlias())
	if err != nil {
		return nil, toGRPCError(err)
	}

	return &pb.Empty{}, nil
}

func (g *grpcServer) GetPaymentPointer(ctx context.Context, req *pb.GetPaymentPointerRequest) (*pb.PaymentPointer, error) {
	w, err := g.b.Wallets().GetFromAddress(ctx, req.Url)
	if err != nil {
		return nil, toGRPCError(err)
	}

	legalName := ""
	kycInfo, err := g.b.KYC().GetIndividualDetails(ctx, w.ID)
	if err == nil {
		legalName = kycInfo.FirstName + " " + kycInfo.LastName
	}

	return &pb.PaymentPointer{
		Url:        w.AddressString(),
		Asset:      "USD",
		AssetScale: int32(2),
		Alias:      w.Name,
		WalletID:   w.ID,
		Formatted:  w.AddressShortString(),
		LegalName:  legalName,
	}, nil
}

func (g *grpcServer) ListWalletPaymentPointers(ctx context.Context, _ *pb.Empty) (*pb.ListWalletPaymentPointersResponse, error) {
	_, err := g.b.Users().UserForContext(ctx)
	if err != nil {
		return nil, UnauthenticatedError("no login found")
	}

	wallet, err := g.b.Wallets().ForContext(ctx)
	if err != nil {
		return nil, ForbiddenError("Unauthenticated.")
	}

	legalName := ""
	kycInfo, err := g.b.KYC().GetIndividualDetails(ctx, wallet.ID)
	if err == nil {
		legalName = kycInfo.FirstName + " " + kycInfo.LastName
	}

	resp := make([]*pb.PaymentPointer, len(wallet.Addresses))
	for i, wa := range wallet.Addresses {

		resp[i] = &pb.PaymentPointer{
			Url:        wa.String(),
			Asset:      "USD",
			AssetScale: int32(0),
			Alias:      wallet.Name,
			WalletID:   wallet.ID,
			Formatted:  wa.ShortString(),
			LegalName:  legalName,
		}
	}

	return &pb.ListWalletPaymentPointersResponse{Pointers: resp}, nil
}

func (g *grpcServer) PaymentPointerExists(ctx context.Context, req *pb.PaymentPointerExistsRequest) (*pb.PaymentPointerExistsResponse, error) {
	exists, err := ops.PaymentPointerExists(ctx, g.b, req.Url)
	if errors.Is(err, wallets.ErrInvalidAddress) {
		return nil, NewValidationError("url", strings.TrimSpace(strings.TrimPrefix(err.Error(), wallets.ErrInvalidAddress.Error())))
	}
	if err != nil {
		return nil, toGRPCError(err)
	}

	return &pb.PaymentPointerExistsResponse{
		Exists: exists,
	}, nil
}
