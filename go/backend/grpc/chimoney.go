package grpc

import (
	"context"
	"errors"

	"gitlab.com/fynbos/backend/currency"

	"gitlab.com/fynbos/backend/providers/chimoney"

	pb "gitlab.com/fynbos/proto/backend/v1"
)

func (s *rpcService) SetChimoneyInterlocEmail(ctx context.Context, req *pb.SetChimoneyInterlocEmailRequest) (*pb.ChimoneyInterlocEmail, error) {
	_, err := s.b.Users().UserForContext(ctx)
	if err != nil {
		return nil, UnauthenticatedError("Unauthenticated.")
	}

	w, err := s.b.Wallets().ForContext(ctx)
	if err != nil {
		return nil, UnauthenticatedError("Unauthenticated.")
	}

	la, err := s.b.Chimoney().AddInterlocEmail(ctx, w.ID, req.Email)
	if err != nil {
		return nil, toGRPCError(err)
	}

	return &pb.ChimoneyInterlocEmail{Email: la.ProviderID}, nil
}

func (s *rpcService) GetChimoneyInterlocEmail(ctx context.Context, _ *pb.Empty) (*pb.ChimoneyInterlocEmail, error) {
	_, err := s.b.Users().UserForContext(ctx)
	if err != nil {
		return nil, UnauthenticatedError("Unauthenticated.")
	}

	w, err := s.b.Wallets().ForContext(ctx)
	if err != nil {
		return nil, UnauthenticatedError("Unauthenticated.")
	}

	email, err := s.b.Chimoney().GetInterlocEmail(ctx, w.ID)
	if err != nil {
		return nil, toGRPCError(err)
	}

	return &pb.ChimoneyInterlocEmail{Email: email}, nil
}

func (s *rpcService) CreateChimoneyWallet(ctx context.Context, _ *pb.Empty) (*pb.Empty, error) {
	_, err := s.b.Users().UserForContext(ctx)
	if err != nil {
		return nil, UnauthenticatedError("Unauthenticated.")
	}

	w, err := s.b.Wallets().ForContext(ctx)
	if err != nil {
		return nil, UnauthenticatedError("Unauthenticated.")
	}

	await, err := s.b.Chimoney().CreateWallet(ctx, w.ID)
	if errors.Is(err, chimoney.ErrNotFound) {
		return nil, FailedPreconditionError("interloc email not found")
	}
	if err != nil {
		return nil, toGRPCError(err)
	}

	var exID string
	err = await(ctx, &exID)
	if err != nil {
		return nil, toGRPCError(err)
	}

	return &pb.Empty{}, nil
}

func (s *rpcService) GetChimoneyDepositLink(ctx context.Context, pbAmt *pb.Amount) (*pb.GetChimoneyDepositLinkResponse, error) {
	_, err := s.b.Users().UserForContext(ctx)
	if err != nil {
		return nil, UnauthenticatedError("Unauthenticated.")
	}

	w, err := s.b.Wallets().ForContext(ctx)
	if err != nil {
		return nil, UnauthenticatedError("Unauthenticated.")
	}

	link, err := s.b.Chimoney().CreateDepositLink(ctx, w.ID, currency.FromPB(pbAmt))
	if err != nil {
		return nil, toGRPCError(err)
	}

	return &pb.GetChimoneyDepositLinkResponse{Link: link}, err
}

func (s *rpcService) CreateChimoneyDeposit(ctx context.Context, req *pb.CreateChimoneyDepositRequest) (*pb.Empty, error) {
	// This method is deprecated and handled via webhooks
	return nil, UnavailableError("CreateChimoneyDeposit is deprecated and handled via webhooks")
}
