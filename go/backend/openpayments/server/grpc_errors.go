package server

import (
	"errors"
	"fmt"

	"gitlab.com/fynbos/backend/openpayments"

	"github.com/go-playground/validator/v10"
	"gitlab.com/fynbos/log"
	"go.uber.org/zap"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var errorStatus = map[error]error{
	openpayments.ErrPaymentPointerNotFound: NotFoundError("payment pointer not found"),
}

func validationDesc(fe validator.FieldError) string {
	switch fe.Tag() {
	case "e164":
		return "Phone number is invalid."
	case "required":
		return "This field is Required."
	case "uuid":
		return "Incorrect format, please provide a UUID."
	case "iso3166_1_alpha2":
		return "Provide a valid country code."
	case "email":
		return "Provide a valid email address."
	case "url":
		return "Provide a valid URL"
	case "iso4217":
		return "Provide a valid currency"
	}

	return ""
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
	return status.Error(codes.Internal, "Internal server error")
}

// ValidationError will build an immutable error representing the status of the response.
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
	return status.Error(codes.Internal, "Internal server error")
}

// NotFoundError will build an immutable error representing the status of the response.
func NotFoundError(message string) error {
	return status.Error(codes.NotFound, "Not found: "+message)
}

func ForbiddenError(message string) error {
	return status.Error(codes.PermissionDenied, "Forbidden: "+message)
}
