package server

import (
	"context"

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

func (g grpcServer) CreatePaymentPointer(ctx context.Context, req *pb.PaymentPointer) (*pb.Empty, error) {

	pp := openpayments.PaymentPointer{
		URL:        req.Url,
		WalletID:   req.WalletID,
		Alias:      req.Alias,
		Asset:      req.Asset,
		AssetScale: int(req.AssetScale),
	}

	err := g.b.Validator().Struct(pp)
	if err != nil {
		return nil, toGRPCError(err)
	}

	err = ops.CreatePaymentPointer(ctx, g.b, pp)
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

	return &pb.PaymentPointer{
		Url:        pp.URL,
		Asset:      pp.Asset,
		AssetScale: int32(pp.AssetScale),
		Alias:      pp.Alias,
		WalletID:   pp.WalletID,
	}, nil
}

func (g grpcServer) ListWalletPaymentPointers(ctx context.Context, req *pb.ListWalletPaymentPointersRequest) (*pb.ListWalletPaymentPointersResponse, error) {
	pp, err := ops.ListWalletPaymentPointers(ctx, g.b, req.WalletID)
	if err != nil {
		return nil, toGRPCError(err)
	}

	resp := make([]*pb.PaymentPointer, len(pp))
	for i, p := range pp {
		resp[i] = &pb.PaymentPointer{
			Url:        p.URL,
			Asset:      p.Asset,
			AssetScale: int32(p.AssetScale),
			Alias:      p.Alias,
			WalletID:   p.WalletID,
		}
	}

	return &pb.ListWalletPaymentPointersResponse{Pointers: resp}, nil
}
