package grpc

import (
	"context"

	"gitlab.com/fynbos/backend/twilio"
	pb "gitlab.com/fynbos/proto/backend/v1"
)

func (s *rpcService) SendOTP(ctx context.Context, _ *pb.Empty) (*pb.Empty, error) {
	user, err := s.userService.ForContext(ctx)
	if err != nil {
		return nil, ForbiddenError("Unauthenticated.")
	}

	id, err := s.identityService.Get(ctx, user.ID)
	if err != nil {
		return nil, grpcError(err)
	}

	_, err = s.twilioService.SendVerificationCode(ctx, id.MobileNumber)
	if err != nil {
		return nil, grpcError(err)
	}

	return &pb.Empty{}, nil
}

func (s *rpcService) ValidateOTP(ctx context.Context, req *pb.ValidateOTPRequest) (*pb.ValidateOTPResponse, error) {
	user, err := s.userService.ForContext(ctx)
	if err != nil {
		return nil, ForbiddenError("Unauthenticated.")
	}

	id, err := s.identityService.Get(ctx, user.ID)
	if err != nil {
		return nil, grpcError(err)
	}

	v, err := s.twilioService.CheckVerificationCode(ctx, &twilio.CheckVerificationCodeArgs{
		PhoneNumber: id.MobileNumber,
		Code:        req.Otp,
	})
	if err != nil {
		return nil, grpcError(err)
	}

	return &pb.ValidateOTPResponse{
		Valid: v.IsValid(),
	}, nil
}
