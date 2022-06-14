package twilio

//go:generate mockgen -destination=./mock.go -package=twilio -source=./service.go

import (
	"context"
	"errors"

	"github.com/go-playground/validator/v10"
	"github.com/twilio/twilio-go"
	"github.com/twilio/twilio-go/client"
	verify "github.com/twilio/twilio-go/rest/verify/v2"
)

type ServiceArgs struct {
	AccountSid   string `validate:"required"`
	AccountToken string `validate:"required"`
	ServiceSid   string `validate:"required"`
	ApiBaseUrl   string // use this to override the default base url
}

var (
	ErrInternal    = errors.New("twilio service: internal error.")
	ErrInvalidCode = errors.New("twilio service: invalid code.")
)

type service struct {
	validator    *validator.Validate
	serviceSid   string
	twilioClient *twilio.RestClient
}

type VerificationStatus struct {
	Sid         string
	Status      string
	PhoneNumber string
}

type Service interface {
	SendVerificationCode(ctx context.Context, phoneNumber string) (*VerificationStatus, error)
	CheckVerificationCode(ctx context.Context, args *CheckVerificationCodeArgs) (*VerificationStatus, error)
}

func NewService(args *ServiceArgs) (Service, error) {
	validator := validator.New()
	err := validator.Struct(args)
	if err != nil {
		return nil, err
	}

	customClient := &CustomClient{
		Client: client.Client{
			Credentials: client.NewCredentials(args.AccountSid, args.AccountToken),
		},
	}

	customClient.SetAccountSid(args.AccountSid)
	customClient.BaseURL = args.ApiBaseUrl

	twilioClient := twilio.NewRestClientWithParams(twilio.ClientParams{
		Client: customClient,
	})

	return &service{
		validator:    validator,
		twilioClient: twilioClient,
		serviceSid:   args.ServiceSid,
	}, nil
}

func (s *service) SendVerificationCode(ctx context.Context, phoneNumber string) (*VerificationStatus, error) {
	params := &verify.CreateVerificationParams{}
	params.SetTo(phoneNumber)
	params.SetChannel("sms")

	res, err := s.twilioClient.VerifyV2.CreateVerification(s.serviceSid, params)
	if err != nil {
		return nil, err
	}

	return &VerificationStatus{
		Sid:         *res.Sid,
		PhoneNumber: *res.To,
		Status:      *res.Status,
	}, nil
}

type CheckVerificationCodeArgs struct {
	PhoneNumber string `validate:"required,e164"`
	Code        string `validate:"required"`
}

func (s *service) CheckVerificationCode(ctx context.Context, args *CheckVerificationCodeArgs) (*VerificationStatus, error) {
	params := &verify.CreateVerificationCheckParams{}
	params.SetTo(args.PhoneNumber)
	params.SetCode(args.Code)

	res, err := s.twilioClient.VerifyV2.CreateVerificationCheck(s.serviceSid, params)
	if err != nil {
		return nil, err
	}

	if *res.Status != "approved" {
		return nil, ErrInvalidCode
	}

	return &VerificationStatus{
		Sid:         *res.Sid,
		PhoneNumber: *res.To,
		Status:      *res.Status,
	}, nil
}
