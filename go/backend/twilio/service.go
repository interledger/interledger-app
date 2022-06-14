package twilio

//go:generate mockgen -destination=./mock.go -package=twilio -source=./service.go

import (
	"context"
	"errors"

	"github.com/go-playground/validator/v10"
	"github.com/twilio/twilio-go"
	openapi "github.com/twilio/twilio-go/rest/verify/v2"
)

type ServiceArgs struct {
	AccountSid	 		 string `validate:"required"`
	AccountToken 		 string `validate:"required"`
	ServiceSid string `validate:"required"`
}

var (
	ErrInternal = errors.New("twilio service: internal error.")
	ErrInvalidCode = errors.New("twilio service: invalid code.")
)

type service struct {
	validator    		 *validator.Validate
	serviceSid string
	twilioClient 		 *twilio.RestClient
}

type Service interface {
	SendVerificationCode(ctx context.Context, phoneNumber string) (*openapi.VerifyV2Verification, error)
	CheckVerificationCode(ctx context.Context, args *CheckVerificationCodeArgs) (*openapi.VerifyV2VerificationCheck, error)
}

func NewService(args *ServiceArgs) (Service, error) {
	validator := validator.New()
	err := validator.Struct(args)
	if err != nil {
		return nil, err
	}

	twilioClient := twilio.NewRestClientWithParams(twilio.ClientParams{
		Username: args.AccountSid,
		Password: args.AccountToken,
	})

	return &service{
		validator: validator,
		twilioClient: twilioClient,
		serviceSid: args.ServiceSid,
	}, nil
}

func (s *service) SendVerificationCode(ctx context.Context, phoneNumber string) (*openapi.VerifyV2Verification, error) {
	params := &openapi.CreateVerificationParams{}
	params.SetTo(phoneNumber)
	params.SetChannel("sms")

	res, err := s.twilioClient.VerifyV2.CreateVerification(s.serviceSid, params)
	if err != nil {
		return nil, err
	}
	
	return res, nil
}

type CheckVerificationCodeArgs struct {
	PhoneNumber string `validate:"required,e164"`
	Code				string `validate:"required"`
}

func (s *service) CheckVerificationCode(ctx context.Context, args *CheckVerificationCodeArgs) (*openapi.VerifyV2VerificationCheck, error) {
	params := &openapi.CreateVerificationCheckParams{}
	params.SetTo(args.PhoneNumber)
	params.SetCode(args.Code)

	res, err := s.twilioClient.VerifyV2.CreateVerificationCheck(s.serviceSid, params)
	if err != nil {
		return nil, err
	}
	
	if *res.Status != "approved" {
		return nil, ErrInvalidCode
	}

	return res, nil
}