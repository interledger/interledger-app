package grpc

import (
	"errors"
	"fmt"

	"github.com/getsentry/sentry-go"

	"gitlab.com/fynbos/backend/errorhandling"
	"gitlab.com/fynbos/backend/kyc"
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
	pb "gitlab.com/fynbos/proto/backend/v1"

	"go.uber.org/zap"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const VALIDATION_ERR_MSG = "Some fields are incorrect."

var errorStatus = map[error]error{
	user.ErrNoUserFound:            newError(codes.Unauthenticated, errorhandling.ErrCodeUserNoUserFound, "Unauthenticated", nil),
	twilio.ErrInvalidOTP:           newValidationErrorSingleField(errorhandling.ErrCodeTwilioInvalidOTP, "OTP", "Could not validate OTP"),
	wallets.ErrDuplicateWallet:     newError(codes.AlreadyExists, errorhandling.ErrCodeWalletsDuplicateWallet, "Wallet already exists", nil),
	wallets.ErrWalletConflict:      newError(codes.FailedPrecondition, errorhandling.ErrCodeWalletsWalletConflict, "Wallet already exists but with different configuration than requested (for example, country, currency, or addresses)", nil),
	linkedaccounts.ErrNotFound:     newError(codes.NotFound, errorhandling.ErrCodeLinkedAccNotFound, "Not found: linked account not found", nil),
	signup.ErrDuplicatePhone:       newError(codes.AlreadyExists, errorhandling.ErrCodeSignupDuplicatePhone, "Phone number already exists with a user.", nil),
	identities.ErrAlreadyExists:    newError(codes.AlreadyExists, errorhandling.ErrCodeIdentitiesAlreadyExists, "Identity already exists", nil),
	wallets.ErrNoWalletFound:       newError(codes.NotFound, errorhandling.ErrCodeWalletsNoWalletFound, "Not found: wallet address not found", nil),
	payments.ErrRequiredActions:    newError(codes.FailedPrecondition, errorhandling.ErrCodePaymentsRequiredActions, "Required details missing for payment", nil),
	payments.ErrInsufficientFunds:  PaymentInsufficientFundsError(),
	kyc.ErrKYCResubmissionRequired: newError(codes.FailedPrecondition, errorhandling.ErrCodeKYCResubmissionRequired, "KYC resubmission required: please update your verification documents", nil),
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
	return newInternalError("Internal server error")
}

func NewValidationError(field string, description string) error {
	st := status.New(codes.InvalidArgument, VALIDATION_ERR_MSG)

	br := &errdetails.BadRequest{}
	appendBrFieldViolation(br, field, description)

	fields := []*pb.AppErrorField{newAppErrorField(field, description)}
	appError := &pb.AppError{
		ErrorCode: errorhandling.ErrCodeValidation,
		Message:   VALIDATION_ERR_MSG,
		Fields:    fields,
	}

	return statusWithDetails(st, br, appError).Err()
}

func NewTwilioError(field string, description string) error {
	st := status.New(codes.InvalidArgument, VALIDATION_ERR_MSG)

	br := &errdetails.BadRequest{}
	appendBrFieldViolation(br, field, description)
	metadata := &errdetails.ErrorInfo{
		Reason: "TwilioError",
	}

	// TODO maybe a TWILIO specific error code?
	fields := []*pb.AppErrorField{newAppErrorField(field, description)}
	appError := &pb.AppError{
		ErrorCode: errorhandling.ErrCodeValidation,
		Message:   VALIDATION_ERR_MSG,
		Fields:    fields,
	}

	return statusWithDetails(st, br, metadata, appError).Err()
}

// Validation error will build an immutable error representing the status of the response.
func ValidationError(err error, description func(validator.FieldError) string) error {
	var validatorError validator.ValidationErrors

	if errors.As(err, &validatorError) {
		st := status.New(codes.InvalidArgument, VALIDATION_ERR_MSG)
		br := &errdetails.BadRequest{}

		fields := []*pb.AppErrorField{}
		for _, err := range validatorError {
			appendBrFieldViolation(br, err.Field(), description(err))
			fields = appendAppErrField(fields, err.Field(), description(err))
		}

		appError := &pb.AppError{
			ErrorCode: errorhandling.ErrCodeValidation,
			Message:   VALIDATION_ERR_MSG,
			Fields:    fields,
		}

		return statusWithDetails(st, br, appError).Err()
	}

	// Default to Internal error if not validation error.
	return newInternalError("Internal server error: Validation error")
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
		return newInternalError("Internal server error: precondition error")
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
		return newInternalError("Internal server error: precondition error")
	}

	return st.Err()
}

func PaymentInsufficientFundsError() error {
	errMsg := "Failed precondition"
	st := status.New(codes.FailedPrecondition, errMsg)
	p := &errdetails.PreconditionFailure{}

	p.Violations = append(p.Violations, &errdetails.PreconditionFailure_Violation{
		Type:        "Payment",
		Subject:     "insufficientFunds",
		Description: "Insufficient Funds",
	})

	appError := &pb.AppError{
		ErrorCode: errorhandling.ErrCodePaymentsInsufficientFunds,
		Message:   errMsg,
	}

	st, err := st.WithDetails(p, appError)
	if err != nil {
		log.Error("failed to encode payment insufficient funds error", zap.Error(err))
		return newInternalError("Internal server error: precondition error")
	}

	return st.Err()
}

func InternalError(message string) error {
	return newInternalError("Internal server error: " + message)
}

func ForbiddenError(message string) error {
	return newError(codes.PermissionDenied, errorhandling.ErrCodeForbidden, "Forbidden: "+message, nil)
}

func UnauthenticatedError(message string) error {
	return newError(codes.Unauthenticated, errorhandling.ErrCodeUnauthorized, "Unauthenticated: "+message, nil)
}

func NotFoundError(message string) error {
	return newError(codes.NotFound, errorhandling.ErrCodeNotFound, "Not found: "+message, nil)
}

func AlreadyExistsError(message string) error {
	return newError(codes.AlreadyExists, errorhandling.ErrCodeConflict, "Already exists: "+message, nil)
}

func FailedPreconditionError(message string) error {
	return newError(codes.FailedPrecondition, errorhandling.ErrCodeBadRequest, "Failed precondition: "+message, nil)
}

func newInternalError(msg string) error {
	return newError(codes.Internal, errorhandling.ErrCodeInternal, msg, nil)
}

func newError(grpcCode codes.Code, appErrCode errorhandling.AppErrorCode, msg string, fields []*pb.AppErrorField) error {
	st := status.New(grpcCode, msg)

	appError := &pb.AppError{
		ErrorCode: appErrCode,
		Message:   msg,
		Fields:    fields,
	}

	return statusWithDetails(st, appError).Err()
}

func newValidationErrorSingleField(appErrCode errorhandling.AppErrorCode, field string, fieldError string) error {
	return newValidationError(appErrCode, []*pb.AppErrorField{newAppErrorField(field, fieldError)})
}

func newValidationError(appErrCode errorhandling.AppErrorCode, fields []*pb.AppErrorField) error {
	st := status.New(codes.InvalidArgument, VALIDATION_ERR_MSG)

	appError := &pb.AppError{
		ErrorCode: appErrCode,
		Message:   VALIDATION_ERR_MSG,
		Fields:    fields,
	}

	return statusWithDetails(st, appError).Err()
}

func newAppErrorField(field string, description string) *pb.AppErrorField {
	return &pb.AppErrorField{
		Field: field,
		Error: description,
	}
}

func appendBrFieldViolation(br *errdetails.BadRequest, field string, description string) {
	v := &errdetails.BadRequest_FieldViolation{
		Field:       field,
		Description: description,
	}
	br.FieldViolations = append(br.FieldViolations, v)
}

func appendAppErrField(fields []*pb.AppErrorField, field string, description string) []*pb.AppErrorField {
	result := append(fields, newAppErrorField(field, description))
	return result
}
