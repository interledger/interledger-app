package grpc

import (
	"context"

	"github.com/interledger/interledger-app/go/backend/twilio"

	pb "github.com/interledger/interledger-app/go/proto/backend/v1"
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

func (s *rpcService) ConfirmUserPhone(ctx context.Context, req *pb.ConfirmUserPhoneRequest) (*pb.Empty, error) {
	u, err := s.b.Users().UserForContext(ctx)
	if err != nil {
		return nil, UnauthenticatedError("Unauthenticated.")
	}

	vc, err := s.b.Twilio().CheckVerificationCode(ctx, &twilio.CheckVerificationCodeArgs{
		PhoneNumber: u.PhoneNumber,
		Code:        req.GetOtp(),
	})
	if err != nil {
		return nil, toGRPCError(err)
	}
	if !vc.IsValid() {
		return nil, NewValidationError("otp", "Invalid OTP")
	}

	if err = s.b.Users().SetPhoneVerified(ctx, u.ID); err != nil {
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
		return nil, NewValidationError("phone", "Phone number is invalid.")
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
		return nil, NewValidationError("phone", "Phone number is invalid.")
	}

	vc, err := s.b.Twilio().CheckVerificationCode(ctx, &twilio.CheckVerificationCodeArgs{
		PhoneNumber: req.GetTo(),
		Code:        req.GetOtp(),
	})
	if err != nil {
		return nil, toGRPCError(err)
	}
	if !vc.IsValid() {
		return nil, NewValidationError("otp", "Invalid OTP")
	}

	return &pb.Empty{}, nil
}
