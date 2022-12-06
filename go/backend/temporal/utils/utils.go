package utils

import (
	"errors"

	"go.temporal.io/api/serviceerror"

	"go.temporal.io/sdk/temporal"
)

func IsNonRetryableError(err error) bool {
	if err == nil {
		return false
	}

	if !temporal.IsApplicationError(err) {
		return false
	}

	var applicationError *temporal.ApplicationError
	if !errors.As(err, &applicationError) {
		return false
	}

	return applicationError.NonRetryable()
}

func IsNotFoundError(err error) bool {
	if err == nil {
		return false
	}

	var notFoundErr *serviceerror.NotFound
	return errors.As(err, &notFoundErr)
}
