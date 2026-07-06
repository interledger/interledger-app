package grpc

import (
	"context"
	"errors"

	"github.com/interledger/interledger-app/go/backend/waitlist"
	pb "github.com/interledger/interledger-app/go/proto/backend/v1"
)

func (s *rpcService) JoinWaitlist(ctx context.Context, req *pb.JoinWaitlistRequest) (*pb.JoinWaitlistResponse, error) {
	err := s.b.Waitlist().Add(ctx, req.Email, req.CountryCode, req.FullName, req.GetMugId(), req.BetaOptIn)
	if errors.Is(err, waitlist.ErrInvalidEmail) {
		return nil, NewValidationError("Email", "Invalid email address.")
	}
	if errors.Is(err, waitlist.ErrInvalidCountry) {
		return nil, NewValidationError("CountryCode", "Invalid country.")
	}
	if errors.Is(err, waitlist.ErrInvalidName) {
		return nil, NewValidationError("FullName", "Full name is required.")
	}
	if err != nil {
		// Do not surface DB connection errors to the user.
		return nil, InternalError("Failed to join waitlist.")
	}

	return &pb.JoinWaitlistResponse{}, nil
}

func (s *rpcService) IsMugAvailable(ctx context.Context, req *pb.IsMugAvailableRequest) (*pb.IsMugAvailableResponse, error) {
	available, err := s.b.Waitlist().IsMugAvailable(ctx, req.MugId)
	if err != nil {
		return nil, toGRPCError(err)
	}

	return &pb.IsMugAvailableResponse{Available: available}, nil
}

func (s *rpcService) CanSignup(ctx context.Context, req *pb.CanSignupRequest) (*pb.CanSignupResponse, error) {
	canSignup, err := s.b.Waitlist().CanSignup(ctx, req.Id)
	if err != nil {
		return nil, InternalError("failed to query can signup.")
	}

	return &pb.CanSignupResponse{
		CanSignup: canSignup,
	}, nil
}

func (s *rpcService) SetSignupComplete(ctx context.Context, req *pb.SetSignupCompleteRequest) (*pb.Empty, error) {
	err := s.b.Waitlist().SetSignupComplete(ctx, req.Id, req.UserId)
	if err != nil {
		return nil, InternalError("failed to query can signup.")
	}

	return &pb.Empty{}, nil
}
