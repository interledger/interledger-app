package admin

import (
	"errors"
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/interledger/interledger-app/go/log"
	"go.uber.org/zap"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var errorStatus = map[error]error{}

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
	log.Error("Unexpected error", zap.Error(err))
	return status.Error(codes.Internal, fmt.Sprintf("Internal server error: %s", err))
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
