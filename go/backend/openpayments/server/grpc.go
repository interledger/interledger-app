package server

import (
	"context"
	"errors"
	"strings"

	"gitlab.com/fynbos/backend/openpayments/workflows"

	"google.golang.org/protobuf/types/known/timestamppb"

	"gitlab.com/fynbos/backend/openpayments"
	"gitlab.com/fynbos/backend/openpayments/ops"
	pb "gitlab.com/fynbos/proto/backend/v1"
)

var _ pb.OpenPaymentServiceServer = &grpcServer{}

type grpcServer struct {
	b Backends
}

func NewGRPCServer(b Backends) pb.OpenPaymentServiceServer {
	return &grpcServer{b: b}
}

func toAmountPB(amount *openpayments.Amount) *pb.Amount {
	if amount == nil {
		return nil
	}
	return &pb.Amount{
		Amount:     amount.Value,
		Asset:      amount.Asset,
		AssetScale: int32(amount.AssetScale),
	}
}

func (g *grpcServer) CreatePaymentPointer(ctx context.Context, req *pb.CreatePaymentPointerRequest) (*pb.Empty, error) {
	_, err := g.b.Users().UserForContext(ctx)
	if err != nil {
		return nil, UnauthenticatedError("no login found")
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

func (g *grpcServer) GetPaymentPointer(ctx context.Context, req *pb.GetPaymentPointerRequest) (*pb.PaymentPointer, error) {
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

func (g *grpcServer) ListWalletPaymentPointers(ctx context.Context, _ *pb.Empty) (*pb.ListWalletPaymentPointersResponse, error) {
	_, err := g.b.Users().UserForContext(ctx)
	if err != nil {
		return nil, UnauthenticatedError("no login found")
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

func (g *grpcServer) PaymentPointerExists(ctx context.Context, req *pb.PaymentPointerExistsRequest) (*pb.PaymentPointerExistsResponse, error) {
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

func (g *grpcServer) CreateQuote(ctx context.Context, req *pb.CreateQuoteRequest) (*pb.Quote, error) {
	args := openpayments.CreateQuoteArgs{
		SendPaymentPointer:    req.SendPaymentPointer,
		ReceivePaymentPointer: req.ReceivePaymentPointer,
		ExpiresAt:             req.ExpiresAt.AsTime(),
		SendAmount: openpayments.Amount{
			Value:      req.GetAmount().GetAmount(),
			Asset:      req.GetAmount().GetAsset(),
			AssetScale: int(req.GetAmount().GetAssetScale()),
		},
	}

	_, err := g.b.Users().UserForContext(ctx)
	if err != nil {
		return nil, UnauthenticatedError("no login found")
	}

	wallet, err := g.b.Users().WalletForContext(ctx)
	if err != nil {
		return nil, ForbiddenError("Unauthenticated.")
	}

	ppl, err := ops.ListWalletPaymentPointers(ctx, g.b, wallet.ID)
	if err != nil {
		return nil, toGRPCError(err)
	}

	var found bool
	for _, pp := range ppl {
		if strings.EqualFold(pp.URL, args.SendPaymentPointer) {
			found = true
		}
	}
	if !found {
		// Signed in user doesn't own the payment pointer it's trying to send from
		return nil, ForbiddenError("Unauthenticated")
	}

	err = g.b.Validator().Struct(args)
	if err != nil {
		return nil, toGRPCError(err)
	}

	q, err := ops.CreateQuote(ctx, g.b, args)
	if err != nil {
		return nil, toGRPCError(err)
	}

	sendAmt := &pb.Amount{
		Amount:     q.SendAmount.Value,
		Asset:      q.SendAmount.Asset,
		AssetScale: int32(q.SendAmount.AssetScale),
	}
	recvAmt := &pb.Amount{
		Amount:     q.ReceiveAmount.Value,
		Asset:      q.ReceiveAmount.Asset,
		AssetScale: int32(q.ReceiveAmount.AssetScale),
	}
	return &pb.Quote{
		Id:             q.ID,
		PaymentPointer: q.PaymentPointer,
		Receiver:       q.IncomingPayment,
		SendAmount:     sendAmt,
		ReceiveAmount:  recvAmt,
		ExpiresAt:      timestamppb.New(q.ExpiresAt),
		CreatedAt:      timestamppb.New(q.CreatedAt),
	}, nil
}

func (g *grpcServer) LookupQuote(ctx context.Context, req *pb.LookupQuoteRequest) (*pb.Quote, error) {
	_, err := g.b.Users().UserForContext(ctx)
	if err != nil {
		return nil, UnauthenticatedError("no login found")
	}

	_, err = g.b.Users().WalletForContext(ctx)
	if err != nil {
		return nil, ForbiddenError("Unauthenticated.")
	}

	q, err := ops.GetQuote(ctx, g.b, req.Id)
	if err != nil {
		return nil, toGRPCError(err)
	}

	sendAmt := toAmountPB(&q.SendAmount)
	recvAmt := toAmountPB(&q.ReceiveAmount)
	return &pb.Quote{
		Id:             q.ID,
		PaymentPointer: q.PaymentPointer,
		Receiver:       q.IncomingPayment,
		SendAmount:     sendAmt,
		ReceiveAmount:  recvAmt,
		ExpiresAt:      timestamppb.New(q.ExpiresAt),
		CreatedAt:      timestamppb.New(q.CreatedAt),
	}, nil
}

func (g *grpcServer) CreateIncomingPayment(ctx context.Context, req *pb.CreateIncomingPaymentRequest) (*pb.IncomingPayment, error) {
	args := openpayments.CreateIncomingPaymentArgs{
		PaymentPointer: req.PaymentPointer,
		ExternalRef:    req.Reference,
		ExpiresAt:      req.ExpiresAt.AsTime(),
	}
	if req.Amount != nil {
		args.IncomingAmount = &openpayments.Amount{
			Value:      req.Amount.Amount,
			Asset:      req.Amount.Asset,
			AssetScale: int(req.Amount.AssetScale),
		}
	}

	_, err := g.b.Users().UserForContext(ctx)
	if err != nil {
		return nil, UnauthenticatedError("no login found")
	}

	wallet, err := g.b.Users().WalletForContext(ctx)
	if err != nil {
		return nil, ForbiddenError("Unauthenticated.")
	}

	ppl, err := ops.ListWalletPaymentPointers(ctx, g.b, wallet.ID)
	if err != nil {
		return nil, toGRPCError(err)
	}

	// TODO: Can you only create incoming payments for yourself...
	// Probably not long term, but for now :shrug:
	var found bool
	for _, pp := range ppl {
		if strings.EqualFold(pp.URL, args.PaymentPointer) {
			found = true
		}
	}
	if !found {
		// Signed in user doesn't own the payment pointer it's trying to send from
		return nil, ForbiddenError("Unauthenticated")
	}

	err = g.b.Validator().Struct(args)
	if err != nil {
		return nil, toGRPCError(err)
	}

	ip, err := ops.CreateIncomingPayment(ctx, g.b, args)
	if err != nil {
		return nil, toGRPCError(err)
	}

	incomingAmt := toAmountPB(ip.IncomingAmount)
	recvAmt := toAmountPB(ip.ReceivedAmount)
	return &pb.IncomingPayment{
		Id:                 ip.ID,
		PaymentPointer:     ip.PaymentPointer,
		FromPaymentPointer: ip.FromPaymentPointer,
		IncomingAmount:     incomingAmt,
		ReceivedAmount:     recvAmt,
		Completed:          ip.Completed,
		ExternalRef:        ip.ExternalRef,
		ExpiresAt:          timestamppb.New(ip.ExpiresAt),
		CreatedAt:          timestamppb.New(ip.CreatedAt),
		UpdatedAt:          timestamppb.New(ip.UpdatedAt),
	}, nil
}

func (g *grpcServer) LookupIncomingPayment(ctx context.Context, req *pb.LookupIncomingPaymentRequest) (*pb.IncomingPayment, error) {
	_, err := g.b.Users().UserForContext(ctx)
	if err != nil {
		return nil, UnauthenticatedError("no login found")
	}

	_, err = g.b.Users().WalletForContext(ctx)
	if err != nil {
		return nil, ForbiddenError("Unauthenticated.")
	}

	ip, err := ops.GetIncomingPayment(ctx, g.b, req.Id)
	if err != nil {
		return nil, toGRPCError(err)
	}

	incomingAmt := toAmountPB(ip.IncomingAmount)
	recvAmt := toAmountPB(ip.ReceivedAmount)
	return &pb.IncomingPayment{
		Id:                 ip.ID,
		PaymentPointer:     ip.PaymentPointer,
		FromPaymentPointer: ip.FromPaymentPointer,
		IncomingAmount:     incomingAmt,
		ReceivedAmount:     recvAmt,
		Completed:          ip.Completed,
		ExternalRef:        ip.ExternalRef,
		ExpiresAt:          timestamppb.New(ip.ExpiresAt),
		CreatedAt:          timestamppb.New(ip.CreatedAt),
		UpdatedAt:          timestamppb.New(ip.UpdatedAt),
	}, nil
}

func (g *grpcServer) CreateOutgoingPayment(ctx context.Context, req *pb.CreateOutgoingPaymentRequest) (*pb.OutgoingPayment, error) {
	args := openpayments.CreateOutgoingPaymentArgs{
		QuoteID:     req.QuoteID,
		Description: req.Description,
		ExternalRef: req.ExternalRef,
		IPAddress:   req.IpAddress,
	}

	err := g.b.Validator().Struct(args)
	if err != nil {
		return nil, toGRPCError(err)
	}

	op, err := workflows.StartOutgoingPayment(ctx, g.b, args)
	if err != nil {
		return nil, toGRPCError(err)
	}

	return &pb.OutgoingPayment{
		Id:               op.ID,
		PaymentPointer:   op.PaymentPointer,
		ToPaymentPointer: op.ToPaymentPointer,
		Failed:           op.Failed,
		Receiver:         op.Receiver,
		SendAmount:       toAmountPB(&op.SendAmount),
		ReceiveAmount:    toAmountPB(&op.ReceiveAmount),
		SentAmount:       toAmountPB(&op.SentAmount),
		Description:      op.Description,
		CreatedAt:        timestamppb.New(op.CreatedAt),
		UpdatedAt:        timestamppb.New(op.UpdatedAt),
	}, nil
}

func (g *grpcServer) LookupOutgoingPayment(ctx context.Context, req *pb.LookupOutgoingPaymentRequest) (*pb.OutgoingPayment, error) {
	op, err := ops.GetOutgoingPayment(ctx, g.b, req.Id)
	if err != nil {
		return nil, toGRPCError(err)
	}

	return &pb.OutgoingPayment{
		Id:               op.ID,
		PaymentPointer:   op.PaymentPointer,
		ToPaymentPointer: op.ToPaymentPointer,
		Failed:           op.Failed,
		Receiver:         op.Receiver,
		SendAmount:       toAmountPB(&op.SendAmount),
		ReceiveAmount:    toAmountPB(&op.ReceiveAmount),
		SentAmount:       toAmountPB(&op.SentAmount),
		Description:      op.Description,
		CreatedAt:        timestamppb.New(op.CreatedAt),
		UpdatedAt:        timestamppb.New(op.UpdatedAt),
	}, nil
}

func (g *grpcServer) ListIncomingPayments(ctx context.Context, req *pb.PaginationRequest) (*pb.ListIncomingPaymentsResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (g *grpcServer) ListOutgoingPayments(ctx context.Context, req *pb.PaginationRequest) (*pb.ListOutgoingPaymentsResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (g *grpcServer) ListPendingOutgoingPayments(ctx context.Context, req *pb.PaginationRequest) (*pb.ListOutgoingPaymentsResponse, error) {
	//TODO implement me
	panic("implement me")
}
