package utils

import (
	"errors"

	"go.temporal.io/api/enums/v1"

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

func IsMaxRetryError(err error) bool {
	if err == nil {
		return false
	}

	var activityError *temporal.ActivityError
	if !errors.As(err, &activityError) {
		return false
	}

	return activityError.RetryState() == enums.RETRY_STATE_MAXIMUM_ATTEMPTS_REACHED || activityError.RetryState() == enums.RETRY_STATE_NON_RETRYABLE_FAILURE
}

func IsNotFoundError(err error) bool {
	if err == nil {
		return false
	}

	var notFoundErr *serviceerror.NotFound
	return errors.As(err, &notFoundErr)
}
