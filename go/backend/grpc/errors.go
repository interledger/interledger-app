package grpc

import (
	"errors"
	"fmt"

	"github.com/getsentry/sentry-go"

	"gitlab.com/fynbos/backend/payments"

	"gitlab.com/fynbos/env"

	"github.com/go-playground/validator/v10"
	"gitlab.com/fynbos/backend/identities"
	"gitlab.com/fynbos/backend/wallets"

	"gitlab.com/fynbos/backend/linkedaccounts"
	"gitlab.com/fynbos/backend/signup"

	"gitlab.com/fynbos/backend/twilio"
	"gitlab.com/fynbos/backend/user"
	"gitlab.com/fynbos/log"
	"go.uber.org/zap"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var errorStatus = map[error]error{
	user.ErrNoUserFound:           status.Error(codes.Unauthenticated, "Unauthenticated"),
	twilio.ErrInvalidOTP:          NewValidationError("OTP", "Could not validate OTP"),
	wallets.ErrDuplicateWallet:    status.Error(codes.AlreadyExists, "Wallet already exists"),
	wallets.ErrWalletConflict:     status.Error(codes.FailedPrecondition, "Wallet already exists but with different configuration than requested (for example, country, currency, or addresses)"),
	linkedaccounts.ErrNotFound:    NotFoundError("linked account not found"),
	signup.ErrDuplicatePhone:      status.Error(codes.AlreadyExists, "Phone number already exists with a user."),
	identities.ErrAlreadyExists:   status.Error(codes.AlreadyExists, "Identity already exists"),
	wallets.ErrNoWalletFound:      NotFoundError("wallet address not found"),
	payments.ErrRequiredActions:   status.Error(codes.FailedPrecondition, "Required details missing for payment"),
	payments.ErrInsufficientFunds: PaymentInsufficientFundsError(),
}

func validationDesc(fe validator.FieldError) string {
	switch fe.Tag() {
	case "e164":
		return "Phone number is invalid."
	case "required":
		return "This field is Required."
	case "uuid":
		return "Incorrect format, please provide a UUID."
	case "len":
		return "Invalid length."
	case "iso3166_1_alpha2":
		return "Provide a valid country code."
	case "iso3166_2":
		return "Provide a valid state code."
	case "email":
		return "Provide a valid email address."
	case "url":
		return "Provide a valid URL"
	case "iso4217":
		return "Provide a valid currency"
	case "ip_addr":
		return "Provide a valid IP address"
	case "gt", "min-number":
		return "Must be greater than " + fe.Param()
	case "min-items":
		return "Must contain at least " + fe.Param()
	case "oneof":
		return fmt.Sprintf("Must be one of [%s]", fe.Param())
	}

	return fe.Error() // default message
}

// toGRPCError converts a given error to its frontend friendly equivalent.
func toGRPCError(err error) error {
	if err == nil {
		return nil
	}

	if !env.IsTest() {
		// This is an info log so that we can omit it easily in production. This should not be a warning
		// because it can be a common occurrence, such as a user not being found. This log is suppressed for
		// unit tests to avoid cluttering the test output.
		log.Info("gRPC error", zap.Error(err))
	}

	// Check if it is a validation error
	var validatorError validator.ValidationErrors
	if errors.As(err, &validatorError) {
		return ValidationError(err, validationDesc)
	}

	// Try for a direct pointer match
	me, ok := errorStatus[err]
	if ok {
		return me
	}

	// In case the error was wrapped.
	for k, v := range errorStatus {
		if errors.Is(err, k) {
			return v
		}
	}

	// Default to a generic error and log
	_ = sentry.CaptureException(err)
	log.Error("Unexpected error", zap.Error(err))
	return status.Error(codes.Internal, "Internal server error")
}

func NewValidationError(field string, description string) error {
	st := status.New(codes.InvalidArgument, "Some fields are incorrect.")
	br := &errdetails.BadRequest{}

	v := &errdetails.BadRequest_FieldViolation{
		Field:       field,
		Description: description,
	}
	br.FieldViolations = append(br.FieldViolations, v)

	st, err := st.WithDetails(br)
	if err != nil {
		// If this errored, it will always error
		// here, so better panic so we can figure
		// out why than have this silently passing.
		panic(fmt.Sprintf("Unexpected error attaching metadata: %v", err))
	}
	return st.Err()
}

func NewTwilioError(field string, description string) error {
	st := status.New(codes.InvalidArgument, "Some fields are incorrect.")
	br := &errdetails.BadRequest{}

	v := &errdetails.BadRequest_FieldViolation{
		Field:       field,
		Description: description,
	}
	br.FieldViolations = append(br.FieldViolations, v)

	metadata := &errdetails.ErrorInfo{
		Reason: "TwilioError",
	}

	st, err := st.WithDetails(br, metadata)
	if err != nil {
		// If this errored, it will always error
		// here, so better panic so we can figure
		// out why than have this silently passing.
		panic(fmt.Sprintf("Unexpected error attaching metadata: %v", err))
	}
	return st.Err()
}

// Validation error will build an immutable error representing the status of the response.
func ValidationError(err error, description func(validator.FieldError) string) error {
	var validatorError validator.ValidationErrors
	if errors.As(err, &validatorError) {
		st := status.New(codes.InvalidArgument, "Some fields are incorrect.")
		br := &errdetails.BadRequest{}

		for _, err := range validatorError {
			v := &errdetails.BadRequest_FieldViolation{
				Field:       err.Field(),
				Description: description(err),
			}
			br.FieldViolations = append(br.FieldViolations, v)
		}

		st, err := st.WithDetails(br)
		if err != nil {
			// If this errored, it will always error
			// here, so better panic so we can figure
			// out why than have this silently passing.
			panic(fmt.Sprintf("Unexpected error attaching metadata: %v", err))
		}
		return st.Err()
	}

	// Default to Internal error if not validation error.
	return status.Error(codes.Internal, "Internal server error: Validation error")
}

func CardPreconditionError(subject, description string) error {
	st := status.New(codes.FailedPrecondition, "Failed precondition")
	p := &errdetails.PreconditionFailure{
		Violations: []*errdetails.PreconditionFailure_Violation{
			{
				Type:        "Card",
				Subject:     subject,
				Description: description,
			},
		},
	}

	st, err := st.WithDetails(p)
	if err != nil {
		log.Error("failed to encode card precondition error", zap.Error(err))
		return status.Error(codes.Internal, "Internal server error: precondition error")
	}

	return st.Err()
}

func PaymentPreconditionError(preconditions []payments.RequiredActionType) error {
	st := status.New(codes.FailedPrecondition, "Failed precondition")
	p := &errdetails.PreconditionFailure{}

	for _, condition := range preconditions {
		v := &errdetails.PreconditionFailure_Violation{
			Type: "Payment",
		}
		switch condition {
		case payments.RequiredActionTypeIPAddress:
			v.Subject = "ipAddress"
			v.Description = "An ip address is required"
		case payments.RequiredActionTypeOTP:
			v.Subject = "otp"
			v.Description = "OTP is required"
		case payments.RequiredActionTypeSenderAmount:
			v.Subject = "senderAmount"
			v.Description = "Amount is required"
		case payments.RequiredActionTypeSenderAccount:
			v.Subject = "senderAccount"
			v.Description = "Account is required"
		case payments.RequiredActionTypeSenderIdentifier:
			v.Subject = "senderIdentifier"
			v.Description = "Sender is required"
		case payments.RequiredActionTypeReceiverAmount:
			v.Subject = "receiverAmount"
			v.Description = "Amount is required"
		case payments.RequiredActionTypeReceiverIdentifier:
			v.Subject = "receiverIdentifier"
			v.Description = "Recipient is required"
		case payments.RequiredActionTypeThreeDS:
			v.Subject = "threeDS"
			v.Description = "3DS is required"
		default:
			log.Error("unknown payment precondition error", zap.String("condition", condition.String()))
			continue
		}

		p.Violations = append(p.Violations, v)
	}

	st, err := st.WithDetails(p)
	if err != nil {
		log.Error("failed to encode payment precondition error", zap.Error(err))
		return status.Error(codes.Internal, "Internal server error: precondition error")
	}

	return st.Err()
}

func PaymentInsufficientFundsError() error {
	st := status.New(codes.FailedPrecondition, "Failed precondition")
	p := &errdetails.PreconditionFailure{}

	p.Violations = append(p.Violations, &errdetails.PreconditionFailure_Violation{
		Type:        "Payment",
		Subject:     "insufficientFunds",
		Description: "Insufficient Funds",
	})

	st, err := st.WithDetails(p)
	if err != nil {
		log.Error("failed to encode payment insufficient funds error", zap.Error(err))
		return status.Error(codes.Internal, "Internal server error: precondition error")
	}

	return st.Err()
}

// Validation error will build an immutable error representing the status of the response.
func InternalError(message string) error {
	return status.Error(codes.Internal, "Internal server error: "+message)
}

// Validation error will build an immutable error representing the status of the response.
func ForbiddenError(message string) error {
	return status.Error(codes.PermissionDenied, "Forbidden: "+message)
}

func UnauthenticatedError(message string) error {
	return status.Error(codes.Unauthenticated, "Unauthenticated: "+message)
}

// Not found error will build an immutable error representing the status of the response.
func NotFoundError(message string) error {
	return status.Error(codes.NotFound, "Not found: "+message)
}

func AlreadyExistsError(message string) error {
	return status.Error(codes.AlreadyExists, "Already exists: "+message)
}

func FailedPreconditionError(message string) error {
	return status.Error(codes.FailedPrecondition, "Failed precondition: "+message)
}

func UnavailableError(message string) error {
	return status.Error(codes.Unavailable, "Unavailable: "+message)
}
