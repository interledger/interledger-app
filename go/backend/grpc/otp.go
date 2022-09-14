package grpc

import (
	"context"

	pb "gitlab.com/fynbos/proto/backend/v1"
)

func (s *rpcService) SendOTP(ctx context.Context, _ *pb.Empty) (*pb.Empty, error) {
	user, err := s.b.Users().ForContext(ctx)
	if err != nil {
		return nil, ForbiddenError("Unauthenticated.")
	}

	id, err := s.b.Identity().Get(ctx, user.ID)
	if err != nil {
		return nil, grpcError(err)
	}

	_, err = s.b.Twilio().SendVerificationCode(ctx, id.MobileNumber)
	if err != nil {
		return nil, grpcError(err)
	}

	return &pb.Empty{}, nil
}
