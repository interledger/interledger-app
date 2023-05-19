package grpc

import (
	"context"

	pb "gitlab.com/fynbos/proto/backend/v1"
)

func (s *rpcService) SendOTP(ctx context.Context, _ *pb.Empty) (*pb.Empty, error) {
	user, err := s.b.Users().UserForContext(ctx)
	if err != nil {
		return nil, UnauthenticatedError("Unauthenticated.")
	}

	_, err = s.b.Twilio().SendVerificationCode(ctx, user.PhoneNumber)
	if err != nil {
		return nil, toGRPCError(err)
	}

	return &pb.Empty{}, nil
}

func (s *rpcService) SendPhoneVerification(
	ctx context.Context,
	req *pb.SendPhoneVerificationRequest,
) (*pb.Empty, error) {
	err := s.b.Validator().VarCtx(ctx, req.To, "required,e164")
	if err != nil {
		return nil, NewValidationError("To", "Phone number is invalid.")
	}

	_, err = s.b.Twilio().SendVerificationCode(ctx, req.To)
	if err != nil {
		return nil, toGRPCError(err)
	}

	return &pb.Empty{}, nil
}
