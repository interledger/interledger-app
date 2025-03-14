package twilio

//go:generate mockgen -destination=./mock.go -package=twilio -source=./service.go

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/twilio/twilio-go"
	"github.com/twilio/twilio-go/client"
	verify "github.com/twilio/twilio-go/rest/verify/v2"
	"gitlab.com/fynbos/env"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

var (
	ErrInvalidArgument = errors.New("twilio error: invalid argument")
	ErrInternal        = errors.New("twilio service: internal error.")
	ErrInvalidOTP      = errors.New("twilio: invalid OTP")
)

const statusApproved = "approved"

type (
	Service interface {
		SendVerificationCode(ctx context.Context, phoneNumber string) (*Verification, error)
		CheckVerificationCode(ctx context.Context, args *CheckVerificationCodeArgs) (*Verification, error)
		ListSuccessfulVerificationAttempts(ctx context.Context, args ListSuccessfulVerificationAttemptsArgs) ([]Verification, error)
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

	Verification struct {
		Sid         string
		PhoneNumber string
		Status      string
		UpdatedAt   time.Time
	}
)

func (v Verification) IsValid() bool {
	return v.Status == statusApproved
}

func NewService(args *ServiceArgs) (Service, error) {
	validator := validator.New()
	err := validator.Struct(args)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", ErrInvalidArgument, err)
	}

	customClient := &CustomClient{
		Client: client.Client{
			Credentials: client.NewCredentials(args.AccountSid, args.AccountToken),
		},
	}

	customClient.SetAccountSid(args.AccountSid)
	customClient.BaseURL = args.ApiBaseUrl
	customClient.HTTPClient = otelhttp.DefaultClient

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
		return nil, fmt.Errorf("%w: %s", ErrInvalidArgument, err)
	}

	if !env.IsProd() && !env.IsSandbox() {
		return &Verification{
			Sid:         "1234",
			PhoneNumber: phoneNumber,
			Status:      "pending",
		}, nil
	}

	params := &verify.CreateVerificationParams{}
	params.SetTo(phoneNumber)
	params.SetChannel("sms")

	res, err := s.twilioClient.VerifyV2.CreateVerification(s.serviceSid, params)
	if err != nil {
		twilioError, ok := err.(*client.TwilioRestError)
		if ok {
			return nil, fmt.Errorf("%w: %s", ErrInternal, twilioError.Message)
		}
		return nil, fmt.Errorf("%w: %s", ErrInternal, err)
	}

	var updatedAt time.Time
	if res.DateUpdated != nil {
		updatedAt = *res.DateUpdated
	}
	return &Verification{
		Sid:         *res.Sid,
		PhoneNumber: *res.To,
		Status:      *res.Status,
		UpdatedAt:   updatedAt,
	}, nil
}

type CheckVerificationCodeArgs struct {
	PhoneNumber    string `validate:"required,e164"`
	Code           string `validate:"required,numeric,len=6"`
	VerificationID string
}

func (s *service) CheckVerificationCode(ctx context.Context, args *CheckVerificationCodeArgs) (*Verification, error) {
	// Short circuit here for environments where there is not twilio integrations
	if !env.IsProd() && !env.IsSandbox() {
		return &Verification{
			Sid:         "1234",
			PhoneNumber: args.PhoneNumber,
			Status:      statusApproved,
		}, nil
	}

	err := s.validator.Struct(args)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", ErrInvalidArgument, err)
	}

	params := &verify.CreateVerificationCheckParams{}
	params.SetTo(args.PhoneNumber)
	params.SetCode(args.Code)
	if args.VerificationID != "" {
		params.SetVerificationSid(args.VerificationID)
	}

	res, err := s.twilioClient.VerifyV2.CreateVerificationCheck(s.serviceSid, params)
	if err != nil {
		twilioError, ok := err.(*client.TwilioRestError)
		if ok {
			return nil, fmt.Errorf("%w: %s", ErrInternal, twilioError.Message)
		}
		return nil, fmt.Errorf("%w: %s", ErrInternal, err)
	}

	return &Verification{
		Sid:         *res.Sid,
		PhoneNumber: *res.To,
		Status:      *res.Status,
	}, nil
}

type ListSuccessfulVerificationAttemptsArgs struct {
	To    string
	Limit int
	After time.Time
}

func (s *service) ListSuccessfulVerificationAttempts(ctx context.Context, args ListSuccessfulVerificationAttemptsArgs) ([]Verification, error) {
	if !env.IsProd() && !env.IsSandbox() {
		return []Verification{
			{
				Sid:         "1234",
				PhoneNumber: args.To,
				Status:      statusApproved,
				UpdatedAt:   time.Now(),
			},
		}, nil
	}

	converted := "converted"
	params := &verify.ListVerificationAttemptParams{
		ChannelDataTo:    &args.To,
		Status:           &converted,
		Limit:            &args.Limit,
		DateCreatedAfter: &args.After,
	}

	res, err := s.twilioClient.VerifyV2.ListVerificationAttempt(params)
	if err != nil {
		twilioError, ok := err.(*client.TwilioRestError)
		if ok {
			return nil, fmt.Errorf("%w: %s", ErrInternal, twilioError.Message)
		}
		return nil, fmt.Errorf("%w: %s", ErrInternal, err)
	}

	var ret []Verification
	for _, ver := range res {
		ret = append(ret, Verification{
			Sid:         getString(ver.Sid),
			PhoneNumber: args.To,
			Status:      statusApproved,
			UpdatedAt:   getTime(ver.DateUpdated),
		})
	}

	return ret, nil
}

func getString(arg *string) string {
	if arg != nil {
		return *arg
	}

	return ""
}

func getTime(arg *time.Time) time.Time {
	if arg != nil {
		return *arg
	}

	return time.Time{}
}
