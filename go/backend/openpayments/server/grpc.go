package server

import (
	"context"

	"gitlab.com/fynbos/backend/grpc"

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
		return nil, grpc.ToGRPCError(err)
	}

	err = ops.CreatePaymentPointer(ctx, g.b, pp)
	if err != nil {
		return nil, grpc.ToGRPCError(err)
	}

	return &pb.Empty{}, nil
}

func (g grpcServer) GetPaymentPointer(ctx context.Context, req *pb.GetPaymentPointerRequest) (*pb.PaymentPointer, error) {
	pp, err := ops.GetPaymentPointer(ctx, g.b, req.Url)
	if err != nil {
		return nil, grpc.ToGRPCError(err)
	}

	return &pb.PaymentPointer{
		Url:        pp.URL,
		Asset:      pp.Asset,
		AssetScale: int32(pp.AssetScale),
		Alias:      pp.Alias,
		WalletID:   pp.WalletID,
	}, nil
}
