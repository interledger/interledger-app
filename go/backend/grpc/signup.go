package grpc

import (
	"context"

	"gitlab.com/fynbos/backend/agreements"
	"gitlab.com/fynbos/backend/signup"
	"gitlab.com/fynbos/log"
	pb "gitlab.com/fynbos/proto/backend/v1"
	"go.uber.org/zap"
)

var signupAgreementIDs []string

func InitAgreementIDs(ids []string) {
	signupAgreementIDs = ids
}

func (s *rpcService) SetSignupUserData(ctx context.Context, req *pb.SetSignupUserDataRequest) (*pb.SetSignupUserDataResponse, error) {
	id := ""
	if req.Id != nil {
		id = *req.Id
	}
	args := signup.UserDataArgs{
		ID:           id,
		FirstName:    req.FirstName,
		LastName:     req.LastName,
		Email:        req.Email,
		CountryCode:  req.CountryCode,
		MobileNumber: req.Mobile,
	}

	err := s.b.Validator().StructCtx(ctx, args)
	if err != nil {
		return nil, toGRPCError(err)
	}

	id, err = s.b.Signup().SetUserData(ctx, args)
	if err != nil {
		return nil, toGRPCError(err)
	}

	return &pb.SetSignupUserDataResponse{Id: id}, nil
}

func (s *rpcService) SetSignupMobileNumber(ctx context.Context, req *pb.SetSignupMobileNumberRequest) (*pb.Empty, error) {
	args := signup.MobileNumberArgs{
		ID:           req.Id,
		MobileNumber: req.Mobile,
		OTP:          req.Otp,
	}
	err := s.b.Validator().StructCtx(ctx, args)
	if err != nil {
		return nil, toGRPCError(err)
	}

	err = s.b.Signup().SetMobileNumber(ctx, args)
	if err != nil {
		return nil, toGRPCError(err)
	}

	return &pb.Empty{}, nil
}

func (s *rpcService) GetSignup(ctx context.Context, req *pb.GetSignupRequest) (*pb.Signup, error) {
	su, err := s.b.Signup().Get(ctx, req.Id)
	if err != nil {
		return nil, toGRPCError(err)
	}

	return &pb.Signup{
		Id:           su.ID,
		FirstName:    su.FirstName,
		LastName:     su.LastName,
		Email:        su.Email,
		CountryCode:  su.CountryCode,
		MobileNumber: su.MobileNumber,
		UserId:       su.UserID,
		Completed:    su.Completed,
	}, nil
}

func (s *rpcService) CompleteSignup(ctx context.Context, req *pb.CompleteSignupRequest) (*pb.Empty, error) {
	err := s.b.Signup().Complete(ctx, req.Id, req.UserId)
	if err != nil {
		return nil, toGRPCError(err)
	}

	agreementIDs := getSignupAgreementIDs()
	if len(agreementIDs) > 0 {
		signErr := s.b.Agreements().Sign(ctx, &agreements.SignArgs{
			AgreementIDs: agreementIDs,
			UserID:       req.UserId,
		})
		if signErr != nil {
			log.Warn("complete_signup: failed to record agreement signatures", zap.Error(signErr), zap.String("userId", req.UserId))
		}
	}

	return &pb.Empty{}, nil
}

func getSignupAgreementIDs() []string {
	return signupAgreementIDs
}
