package twilio

//go:generate mockgen -destination=./mock.go -package=twilio -source=./service.go

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/interledger/interledger-app/go/log"
	"github.com/twilio/twilio-go"
	"github.com/twilio/twilio-go/client"
	verify "github.com/twilio/twilio-go/rest/verify/v2"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.uber.org/zap"
)

var twilioRequestTimeout = 10 * time.Second

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
		Enabled      bool   // when false, all methods return stub responses without calling Twilio
	}

	service struct {
		validator    *validator.Validate
		serviceSid   string
		twilioClient *twilio.RestClient
		enabled      bool
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
	if args == nil {
		return nil, fmt.Errorf("%w: args must not be nil", ErrInvalidArgument)
	}
	if args.Enabled {
		v := validator.New()
		if err := v.Struct(args); err != nil {
			return nil, fmt.Errorf("%w: %s", ErrInvalidArgument, err)
		}
	}

	customClient := &CustomClient{
		Client: client.Client{
			Credentials: client.NewCredentials(args.AccountSid, args.AccountToken),
		},
	}

	customClient.SetAccountSid(args.AccountSid)
	customClient.BaseURL = args.ApiBaseUrl
	customClient.HTTPClient = cloneHTTPClientWithTimeout(otelhttp.DefaultClient, twilioRequestTimeout)

	twilioClient := twilio.NewRestClientWithParams(twilio.ClientParams{
		Client: customClient,
	})

	if args.Enabled {
		if err := validateConfiguration(twilioClient, args.ServiceSid); err != nil {
			return nil, err
		}
	}

	return &service{
		validator:    validator.New(),
		twilioClient: twilioClient,
		serviceSid:   args.ServiceSid,
		enabled:      args.Enabled,
	}, nil
}

func cloneHTTPClientWithTimeout(base *http.Client, timeout time.Duration) *http.Client {
	if base == nil {
		return &http.Client{Timeout: timeout}
	}

	cloned := *base
	cloned.Timeout = timeout

	return &cloned
}

func validateConfiguration(twilioClient *twilio.RestClient, serviceSid string) error {
	_, err := twilioClient.VerifyV2.FetchService(serviceSid)
	if err == nil {
		return nil
	}

	var twilioError *client.TwilioRestError
	if errors.As(err, &twilioError) {
		return fmt.Errorf("%w: twilio verify service validation failed (code=%d): %s", ErrInvalidArgument, twilioError.Code, twilioError.Message)
	}

	return fmt.Errorf("%w: twilio verify service validation failed: %s", ErrInternal, err)
}

func (s *service) SendVerificationCode(ctx context.Context, phoneNumber string) (*Verification, error) {
	err := s.validator.Var(phoneNumber, "required,e164")
	if err != nil {
		return nil, fmt.Errorf("%w: %s", ErrInvalidArgument, err)
	}

	if !s.enabled {
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
	if !s.enabled {
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
			switch twilioError.Code {
			case 60202:
				log.Warn("Invalid verification code: ", zap.String("message", twilioError.Message))
				return nil, fmt.Errorf("%w: %s", ErrInvalidOTP, "Invalid verification code")
			case 60203:
				log.Warn("Maximum verification attempts reached: ", zap.String("message", twilioError.Message))
				return nil, fmt.Errorf("%w: %s", ErrInvalidOTP, "Maximum verification attempts reached")
			case 60200:
				log.Warn("Invalid phone number format: ", zap.String("message", twilioError.Message))
				return nil, fmt.Errorf("%w: %s", ErrInvalidArgument, "Invalid phone number format")
			default:
				log.Warn("Twilio verification error", zap.Int("code", twilioError.Code), zap.String("message", twilioError.Message))
				return nil, fmt.Errorf("%w: %s", ErrInternal, twilioError.Message)
			}
		}
		log.Warn("Code verification error", zap.Error(err))
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
	if !s.enabled {
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
