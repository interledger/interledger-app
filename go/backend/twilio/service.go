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

var (
	ErrInternal    = errors.New("twilio service: internal error.")
	ErrInvalidCode = errors.New("twilio service: invalid code.")
)

type (
	Service interface {
		SendVerificationCode(ctx context.Context, phoneNumber string) (*Verification, error)
		CheckVerificationCode(ctx context.Context, args *CheckVerificationCodeArgs) (*Verification, error)
	}

	ServiceArgs struct {
		AccountSid   string `validate:"required"`
		AccountToken string `validate:"required"`
		ServiceSid   string `validate:"required"`
		ApiBaseUrl   string // use this to override the default base url
	}

	service struct {
		validator    *validator.Validate
		serviceSid   string
		twilioClient *twilio.RestClient
	}

	RateLimits struct {
		PhoneNumberVerificationSendTimeout string `json:"phone_number_verification_send_timeout"` // this is the unique identifier of the service rate limit
	}

	Verification struct {
		Sid         string
		PhoneNumber string
		Status      string
	}
)

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

func (s *service) SendVerificationCode(ctx context.Context, phoneNumber string) (*Verification, error) {
	err := s.validator.Var(phoneNumber, "required,e164")
	if err != nil {
		return nil, err
	}

	params := &verify.CreateVerificationParams{}
	params.SetTo(phoneNumber)
	params.SetChannel("sms")
	params.SetRateLimits(&RateLimits{
		PhoneNumberVerificationSendTimeout: phoneNumber,
	})
	// TODO: add built-in timeout

	res, err := s.twilioClient.VerifyV2.CreateVerification(s.serviceSid, params)
	if err != nil {
		return nil, err
	}

	return &Verification{
		Sid:         *res.Sid,
		PhoneNumber: *res.To,
		Status:      *res.Status,
	}, nil
}

type CheckVerificationCodeArgs struct {
	PhoneNumber string `validate:"required,e164"`
	Code 				string `validate:"required,numeric,len=6"`
}

func (s *service) CheckVerificationCode(ctx context.Context, args *CheckVerificationCodeArgs) (*Verification, error) {
	err := s.validator.Struct(args)
	if err != nil {
		return nil, err
	}

	params := &verify.CreateVerificationCheckParams{}
	params.SetTo(args.PhoneNumber)
	params.SetCode(args.Code)

	res, err := s.twilioClient.VerifyV2.CreateVerificationCheck(s.serviceSid, params)
	if err != nil {
		return nil, err
	}

	return &Verification{
		Sid:         *res.Sid,
		PhoneNumber: *res.To,
		Status:      *res.Status,
	}, nil
}
