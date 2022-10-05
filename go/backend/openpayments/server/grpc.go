package server

import (
	"context"
	"errors"
	"strings"

	"gitlab.com/fynbos/backend/openpayments"
	"gitlab.com/fynbos/backend/openpayments/ops"
	pb "gitlab.com/fynbos/proto/backend/v1"
)

var _ pb.OpenPaymentServiceServer = grpcServer{}

type grpcServer struct {
	b Backends
}

func NewGRPCServer(b Backends) pb.OpenPaymentServiceServer {
	return &grpcServer{b: b}
}

func (g grpcServer) CreatePaymentPointer(ctx context.Context, req *pb.CreatePaymentPointerRequest) (*pb.Empty, error) {
	_, err := g.b.Users().UserForContext(ctx)
	if err != nil {
		return nil, ForbiddenError("Unauthenticated.")
	}

	wallet, err := g.b.Users().WalletForContext(ctx)
	if err != nil {
		return nil, ForbiddenError("Unauthenticated.")
	}

	pp := openpayments.PaymentPointer{
		URL:        req.Url,
		WalletID:   wallet.ID,
		Alias:      req.Alias,
		Asset:      req.Asset,
		AssetScale: int(req.AssetScale),
	}

	err = g.b.Validator().Struct(pp)
	if err != nil {
		return nil, toGRPCError(err)
	}

	err = ops.CreatePaymentPointer(ctx, g.b, pp)
	if errors.Is(err, openpayments.ErrInvalidPointerPath) {
		return nil, NewValidationError("url", strings.TrimSpace(strings.TrimPrefix(err.Error(), openpayments.ErrInvalidPointerPath.Error())))
	}
	if err != nil {
		return nil, toGRPCError(err)
	}

	return &pb.Empty{}, nil
}

func (g grpcServer) GetPaymentPointer(ctx context.Context, req *pb.GetPaymentPointerRequest) (*pb.PaymentPointer, error) {
	pp, err := ops.GetPaymentPointer(ctx, g.b, req.Url)
	if err != nil {
		return nil, toGRPCError(err)
	}

	formattedPP, err := ops.FormattedPaymentPointer(pp.URL)
	if err != nil {
		return nil, toGRPCError(err)
	}

	return &pb.PaymentPointer{
		Url:        pp.URL,
		Asset:      pp.Asset,
		AssetScale: int32(pp.AssetScale),
		Alias:      pp.Alias,
		WalletID:   pp.WalletID,
		Formatted:  formattedPP,
	}, nil
}

func (g grpcServer) ListWalletPaymentPointers(ctx context.Context, _ *pb.Empty) (*pb.ListWalletPaymentPointersResponse, error) {
	_, err := g.b.Users().UserForContext(ctx)
	if err != nil {
		return nil, ForbiddenError("Unauthenticated.")
	}

	wallet, err := g.b.Users().WalletForContext(ctx)
	if err != nil {
		return nil, ForbiddenError("Unauthenticated.")
	}

	pp, err := ops.ListWalletPaymentPointers(ctx, g.b, wallet.ID)
	if err != nil {
		return nil, toGRPCError(err)
	}

	resp := make([]*pb.PaymentPointer, len(pp))
	for i, p := range pp {
		formattedPP, err := ops.FormattedPaymentPointer(p.URL)
		if err != nil {
			return nil, toGRPCError(err)
		}
		resp[i] = &pb.PaymentPointer{
			Url:        p.URL,
			Asset:      p.Asset,
			AssetScale: int32(p.AssetScale),
			Alias:      p.Alias,
			WalletID:   p.WalletID,
			Formatted:  formattedPP,
		}
	}

	return &pb.ListWalletPaymentPointersResponse{Pointers: resp}, nil
}

func (g grpcServer) PaymentPointerExists(ctx context.Context, req *pb.PaymentPointerExistsRequest) (*pb.PaymentPointerExistsResponse, error) {
	exists, err := ops.PaymentPointerExists(ctx, g.b, req.Url)
	if errors.Is(err, openpayments.ErrInvalidPointerPath) {
		return nil, NewValidationError("url", strings.TrimSpace(strings.TrimPrefix(err.Error(), openpayments.ErrInvalidPointerPath.Error())))
	}
	if err != nil {
		return nil, toGRPCError(err)
	}

	return &pb.PaymentPointerExistsResponse{
		Exists: exists,
	}, nil
}
