package utils

import (
	"errors"

	"go.temporal.io/sdk/temporal"
)

func IsNonRetryableError(err error) bool {
	if err == nil {
		return false
	}

	// Cancelled errors cannot be recovered from
	if temporal.IsCanceledError(err) {
		return true
	}

	if !temporal.IsApplicationError(err) {
		return false
	}

	var applicationError *temporal.ApplicationError
	errors.As(err, &applicationError)

	return applicationError.NonRetryable()
}
