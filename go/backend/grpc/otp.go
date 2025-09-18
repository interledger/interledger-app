package grpc

import (
	"context"
	"errors"

	"gitlab.com/fynbos/backend/twilio"

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

func (s *rpcService) CheckPhoneVerification(
	ctx context.Context,
	req *pb.CheckPhoneVerificationRequest,
) (*pb.Empty, error) {
	err := s.b.Validator().VarCtx(ctx, req.To, "required,e164")
	if err != nil {
		return nil, NewValidationError("To", "Phone number is invalid.")
	}

	vc, err := s.b.Twilio().CheckVerificationCode(ctx, &twilio.CheckVerificationCodeArgs{
		PhoneNumber: req.GetTo(),
		Code:        req.GetOtp(),
	})
	if err != nil {
		if errors.Is(err, twilio.ErrInvalidOTP) {
			return nil, NewTwilioError("Code", "Invalid verification code")
		}
		if errors.Is(err, twilio.ErrInvalidArgument) {
			return nil, NewTwilioError("To", "Invalid phone number format")
		}
		return nil, toGRPCError(err)
	}
	if !vc.IsValid() {
		return nil, NewValidationError("otp", "Invalid OTP")
	}

	return &pb.Empty{}, nil
}
