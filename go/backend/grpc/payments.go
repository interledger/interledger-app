package grpc

import (
	"context"

	"gitlab.com/fynbos/backend/payments"
	pb "gitlab.com/fynbos/proto/backend/v1"
)

func (s *rpcService) InitiateOutgoingPayment(ctx context.Context, req *pb.InitiateOutgoingPaymentRequest) (*pb.InitiateOutgoingPaymentResponse, error) {
	user, err := s.b.Users().UserForContext(ctx)
	if err != nil {
		return nil, ForbiddenError("Unauthenticated.")
	}

	p, err := s.b.Payments().InitiateOutgoingPayment(ctx, payments.InitiateOutgoingPaymentArgs{
		Amount: req.Amount,
		To:     req.To,
		OTP:    req.Otp,
		UserID: user.ID,
	})
	if err != nil {
		return nil, grpcError(err)
	}

	return &pb.InitiateOutgoingPaymentResponse{Id: p.ID}, nil
}

func (s *rpcService) GetOutgoingPayment(ctx context.Context, req *pb.GetOutgoingPaymentRequest) (*pb.OutgoingPayment, error) {
	user, err := s.b.Users().UserForContext(ctx)
	if err != nil {
		return nil, ForbiddenError("Unauthenticated.")
	}

	p, err := s.b.Payments().Get(ctx, req.Id, user.ID)
	if err != nil {
		return nil, grpcError(err)
	}

	return &pb.OutgoingPayment{
		Id:          p.ID,
		AccountId:   p.AccountID,
		Destination: p.Destination,
		Amount:      p.Amount,
		State:       p.State.String(),
	}, nil
}
