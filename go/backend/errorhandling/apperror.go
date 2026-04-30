package errorhandling

import (
	"errors"
	"fmt"

	"github.com/go-playground/validator/v10"
	"gitlab.com/fynbos/backend/identities"
	"gitlab.com/fynbos/backend/kyc"
	"gitlab.com/fynbos/backend/linkedaccounts"
	"gitlab.com/fynbos/backend/payments"
	"gitlab.com/fynbos/backend/signup"
	"gitlab.com/fynbos/backend/twilio"
	"gitlab.com/fynbos/backend/user"
	"gitlab.com/fynbos/backend/wallets"
)

const VALIDATION_ERR_MSG = "Some fields are incorrect."

var appErrors = map[error]*AppError{
	user.ErrNoUserFound:            newError(ErrCodeUnauthorized, "Unauthenticated", nil, user.ErrNoUserFound),
	twilio.ErrInvalidOTP:           newErrorSingleField(ErrCodeTwilioInvalidOTP, VALIDATION_ERR_MSG, "OTP", "Could not validate OTP", twilio.ErrInvalidOTP),
	wallets.ErrDuplicateWallet:     newError(ErrCodeWalletsDuplicateWallet, "Wallet already exists", nil, wallets.ErrDuplicateWallet),
	wallets.ErrWalletConflict:      newError(ErrCodeWalletsWalletConflict, "Wallet already exists but with different configuration than requested (for example, country, currency, or addresses)", nil, wallets.ErrWalletConflict),
	linkedaccounts.ErrNotFound:     newError(ErrCodeLinkedAccNotFound, "Not found: linked account not found", nil, linkedaccounts.ErrNotFound),
	signup.ErrDuplicatePhone:       newError(ErrCodeSignupDuplicatePhone, "Phone number already exists with a user.", nil, signup.ErrDuplicatePhone),
	identities.ErrAlreadyExists:    newError(ErrCodeIdentitiesAlreadyExists, "Identity already exists", nil, identities.ErrAlreadyExists),
	wallets.ErrNoWalletFound:       newError(ErrCodeWalletsNoWalletFound, "Not found: wallet address not found", nil, wallets.ErrNoWalletFound),
	payments.ErrRequiredActions:    newError(ErrCodePaymentsRequiredActions, "Required details missing for payment", nil, payments.ErrRequiredActions),
	payments.ErrInsufficientFunds:  newError(ErrCodePaymentsInsufficientFunds, "Insufficient Funds", nil, payments.ErrInsufficientFunds),
	kyc.ErrKYCResubmissionRequired: newError(ErrCodeKYCResubmissionRequired, "KYC resubmission required: please update your verification documents", nil, kyc.ErrKYCResubmissionRequired),
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

type AppError struct {
	ErrorCode string
	Message   string
	Fields    []AppErrorField
	Cause     error
}

type AppErrorField struct {
	Field string
	Error string
}

// TODO check
func (e *AppError) Error() string {
	return fmt.Sprintf("AppError: code=%s msg=%s fields=%v cause=%v", e.ErrorCode, e.Message, e.Fields, e.Cause)
}

func (e *AppError) Unwrap() error {
	return e.Cause
}

// func NewInternalError(msg string) *AppError {
// 	return &AppError{
// 		ErrorCode: ErrCodeInternal,
// 		Message:   msg,
// 	}
// }

// ToAppError converts a given error-like object that can be converted to http response or grpc response.
func ToAppError(err error) *AppError {
	if err == nil {
		return nil
	}

	var appError *AppError
	if errors.As(err, &appError) {
		return appError
	}

	// Check if it is a validation error
	var validatorError validator.ValidationErrors
	if errors.As(err, &validatorError) {
		return validationError(err, validationDesc)
	}

	// Try for a direct pointer match
	me, ok := appErrors[err]
	if ok {
		return me
	}

	// In case the error was wrapped.
	for k, v := range appErrors {
		if errors.Is(err, k) {
			return v
		}
	}

	return newInternalError("Internal server error")
}

// func NewBadRequestError(msg string) *AppError {
// 	return &AppError{
// 		StatusCode: StatusCodeBadRequest,
// 		ErrorCode:  ErrCodeBadRequest,
// 		Message:    msg,
// 	}
// }

func newError(errCode AppErrorCode, msg string, fields []AppErrorField, err error) *AppError {
	appError := &AppError{
		ErrorCode: errCode,
		Message:   msg,
		Fields:    fields,
		Cause:     err,
	}

	return appError
}

func newErrorSingleField(errCode AppErrorCode, msg string, field string, fieldError string, err error) *AppError {
	return newError(errCode, msg, []AppErrorField{newAppErrorField(field, fieldError)}, err)
}

func appendAppErrField(fields []AppErrorField, field string, description string) []AppErrorField {
	result := append(fields, newAppErrorField(field, description))
	return result
}

func newAppErrorField(field string, description string) AppErrorField {
	return AppErrorField{
		Field: field,
		Error: description,
	}
}

func newInternalError(msg string) *AppError {
	return newError(ErrCodeInternal, msg, nil, nil)
}

func validationError(err error, description func(validator.FieldError) string) *AppError {
	var validatorError validator.ValidationErrors

	if errors.As(err, &validatorError) {
		fields := []AppErrorField{}
		for _, err := range validatorError {
			fields = appendAppErrField(fields, err.Field(), description(err))
		}

		return newError(ErrCodeValidation, VALIDATION_ERR_MSG, fields, err)
	}

	// Default to Internal error if not validation error.
	return newInternalError("Internal server error: Validation error")
}

// func NewValidationError(field string, description string) *AppError {
// 	fields := []AppErrorField{newAppErrorField(field, description)}

// 	return newError(ErrCodeValidation, VALIDATION_ERR_MSG, fields, nil)
// }
