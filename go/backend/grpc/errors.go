package grpc

import (
	"errors"
	"fmt"

	"github.com/go-playground/validator/v10"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

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

// Not found error will build an immutable error representing the status of the response.
func NotFoundError(message string) error {
	return status.Error(codes.NotFound, "Not found: "+message)
}
