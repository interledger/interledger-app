package grpc

import (
	"context"
	"errors"

	"gitlab.com/fynbos/backend/waitlist"
	pb "gitlab.com/fynbos/proto/backend/v1"
)

func (s *rpcService) JoinWaitlist(ctx context.Context, req *pb.JoinWaitlistRequest) (*pb.JoinWaitlistResponse, error) {
	err := s.b.Waitlist().Add(ctx, req.Email, req.CountryCode, req.FullName)
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
