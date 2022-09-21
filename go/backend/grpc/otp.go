package grpc

import (
	"context"

	pb "gitlab.com/fynbos/proto/backend/v1"
)

func (s *rpcService) SendOTP(ctx context.Context, _ *pb.Empty) (*pb.Empty, error) {
	user, err := s.b.Users().UserForContext(ctx)
	if err != nil {
		return nil, ForbiddenError("Unauthenticated.")
	}

	_, err = s.b.Twilio().SendVerificationCode(ctx, user.PhoneNumber)
	if err != nil {
		return nil, grpcError(err)
	}

	return &pb.Empty{}, nil
}
