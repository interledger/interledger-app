package grpc

import (
	"errors"
	"fmt"
	"reflect"
	"testing"

	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/fynbos/backend/errcodes"
	"gitlab.com/fynbos/backend/identities"
	"gitlab.com/fynbos/backend/kyc"
	"gitlab.com/fynbos/backend/linkedaccounts"
	"gitlab.com/fynbos/backend/payments"
	"gitlab.com/fynbos/backend/signup"
	"gitlab.com/fynbos/backend/twilio"
	"gitlab.com/fynbos/backend/user"
	"gitlab.com/fynbos/backend/wallets"
	pb "gitlab.com/fynbos/proto/backend/v1"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// stubFieldError is a minimal implementation of validator.FieldError for testing.
type stubFieldError struct {
	tag    string
	param  string
	errMsg string
	field  string
}

func (s stubFieldError) Tag() string                      { return s.tag }
func (s stubFieldError) ActualTag() string                { return s.tag }
func (s stubFieldError) Namespace() string                { return "" }
func (s stubFieldError) StructNamespace() string          { return "" }
func (s stubFieldError) Field() string                    { return s.field }
func (s stubFieldError) StructField() string              { return s.field }
func (s stubFieldError) Value() any                       { return nil }
func (s stubFieldError) Param() string                    { return s.param }
func (s stubFieldError) Kind() reflect.Kind               { return reflect.String }
func (s stubFieldError) Type() reflect.Type               { return reflect.TypeFor[string]() }
func (s stubFieldError) Translate(_ ut.Translator) string { return "" }
func (s stubFieldError) Error() string                    { return s.errMsg }

func TestValidationDesc(t *testing.T) {
	t.Parallel()

	tests := []struct {
		tag      string
		param    string
		errMsg   string
		expected string
	}{
		{"e164", "", "", "Phone number is invalid."},
		{"required", "", "", "This field is Required."},
		{"uuid", "", "", "Incorrect format, please provide a UUID."},
		{"len", "", "", "Invalid length."},
		{"iso3166_1_alpha2", "", "", "Provide a valid country code."},
		{"iso3166_2", "", "", "Provide a valid state code."},
		{"email", "", "", "Provide a valid email address."},
		{"url", "", "", "Provide a valid URL"},
		{"iso4217", "", "", "Provide a valid currency"},
		{"ip_addr", "", "", "Provide a valid IP address"},
		{"gt", "0", "", "Must be greater than 0"},
		{"min-number", "5", "", "Must be greater than 5"},
		{"min-items", "3", "", "Must contain at least 3"},
		{"oneof", "a b c", "", "Must be one of [a b c]"},
		{"unknown", "", "fallback error message", "fallback error message"},
	}

	for _, tc := range tests {
		t.Run(tc.tag, func(t *testing.T) {
			t.Parallel()
			fe := stubFieldError{tag: tc.tag, param: tc.param, errMsg: tc.errMsg}
			assert.Equal(t, tc.expected, validationDesc(fe))
		})
	}
}

func TestToGRPCError(t *testing.T) {
	t.Parallel()

	t.Run("nil error returns nil", func(t *testing.T) {
		t.Parallel()
		assert.Nil(t, toGRPCError(nil))
	})

	t.Run("validation error returns proper error codes", func(t *testing.T) {
		t.Parallel()
		v := validator.New()
		type S struct {
			Email string `validate:"required,email"`
		}
		validationErr := v.Struct(S{})
		require.Error(t, validationErr)

		result := toGRPCError(validationErr)
		st, ok := status.FromError(result)
		require.True(t, ok)
		assert.Equal(t, codes.InvalidArgument, st.Code())

		appErr := statusFindDetail[*pb.AppError](st)
		require.NotNil(t, appErr)
		assert.Equal(t, errcodes.ErrCodeValidation, appErr.ErrorCode)
	})

	t.Run("known error maps to correct gRPC code and AppError", func(t *testing.T) {
		t.Parallel()
		tests := []struct {
			name            string
			err             error
			expectedCode    codes.Code
			expectedAppCode errcodes.AppErrorCode
		}{
			{"user not found", user.ErrNoUserFound, codes.Unauthenticated, errcodes.ErrCodeUserNoUserFound},
			{"duplicate wallet", wallets.ErrDuplicateWallet, codes.AlreadyExists, errcodes.ErrCodeWalletsDuplicateWallet},
			{"wallet conflict", wallets.ErrWalletConflict, codes.FailedPrecondition, errcodes.ErrCodeWalletsWalletConflict},
			{"linked account not found", linkedaccounts.ErrNotFound, codes.NotFound, errcodes.ErrCodeLinkedAccNotFound},
			{"duplicate phone", signup.ErrDuplicatePhone, codes.AlreadyExists, errcodes.ErrCodeSignupDuplicatePhone},
			{"identity already exists", identities.ErrAlreadyExists, codes.AlreadyExists, errcodes.ErrCodeIdentitiesAlreadyExists},
			{"no wallet found", wallets.ErrNoWalletFound, codes.NotFound, errcodes.ErrCodeWalletsNoWalletFound},
			{"payments required actions", payments.ErrRequiredActions, codes.FailedPrecondition, errcodes.ErrCodePaymentsRequiredActions},
			{"payments insufficient funds", payments.ErrInsufficientFunds, codes.FailedPrecondition, errcodes.ErrCodePaymentsInsufficientFunds},
			{"kyc resubmission required", kyc.ErrKYCResubmissionRequired, codes.FailedPrecondition, errcodes.ErrCodeKYCResubmissionRequired},
			{"twilio invalid otp", twilio.ErrInvalidOTP, codes.InvalidArgument, errcodes.ErrCodeTwilioInvalidOTP},
		}

		for _, tc := range tests {
			t.Run(tc.name, func(t *testing.T) {
				t.Parallel()
				result := toGRPCError(tc.err)
				st, ok := status.FromError(result)
				require.True(t, ok)
				assert.Equal(t, tc.expectedCode, st.Code())

				appErr := statusFindDetail[*pb.AppError](st)
				require.NotNil(t, appErr)
				assert.Equal(t, tc.expectedAppCode, appErr.ErrorCode)
			})
		}
	})

	t.Run("wrapped known error maps correctly", func(t *testing.T) {
		t.Parallel()
		wrapped := fmt.Errorf("context: %w", user.ErrNoUserFound)
		result := toGRPCError(wrapped)
		st, ok := status.FromError(result)
		require.True(t, ok)
		assert.Equal(t, codes.Unauthenticated, st.Code())

		appErr := statusFindDetail[*pb.AppError](st)
		require.NotNil(t, appErr)
		assert.Equal(t, errcodes.ErrCodeUserNoUserFound, appErr.ErrorCode)
	})

	t.Run("unknown error returns Internal", func(t *testing.T) {
		t.Parallel()
		result := toGRPCError(errors.New("some unexpected error"))
		st, ok := status.FromError(result)
		require.True(t, ok)
		assert.Equal(t, codes.Internal, st.Code())

		appErr := statusFindDetail[*pb.AppError](st)
		require.NotNil(t, appErr)
		assert.Equal(t, errcodes.ErrCodeInternal, appErr.ErrorCode)
	})
}

func TestNewValidationError(t *testing.T) {
	t.Parallel()

	err := NewValidationError("email", "Provide a valid email address.")
	st, ok := status.FromError(err)
	require.True(t, ok)
	assert.Equal(t, codes.InvalidArgument, st.Code())
	assert.Equal(t, VALIDATION_ERR_MSG, st.Message())

	br := statusFindDetail[*errdetails.BadRequest](st)
	require.NotNil(t, br)
	require.Len(t, br.FieldViolations, 1)
	assert.Equal(t, "email", br.FieldViolations[0].Field)
	assert.Equal(t, "Provide a valid email address.", br.FieldViolations[0].Description)

	appErr := statusFindDetail[*pb.AppError](st)
	require.NotNil(t, appErr)
	assert.Equal(t, errcodes.ErrCodeValidation, appErr.ErrorCode)
	require.Len(t, appErr.Fields, 1)
	assert.Equal(t, "email", appErr.Fields[0].Field)
}

func TestNewTwilioError(t *testing.T) {
	t.Parallel()

	err := NewTwilioError("otp", "Invalid OTP")
	st, ok := status.FromError(err)
	require.True(t, ok)
	assert.Equal(t, codes.InvalidArgument, st.Code())

	br := statusFindDetail[*errdetails.BadRequest](st)
	require.NotNil(t, br)
	require.Len(t, br.FieldViolations, 1)
	assert.Equal(t, "otp", br.FieldViolations[0].Field)

	errInfo := statusFindDetail[*errdetails.ErrorInfo](st)
	require.NotNil(t, errInfo)
	assert.Equal(t, "TwilioError", errInfo.Reason)

	appErr := statusFindDetail[*pb.AppError](st)
	require.NotNil(t, appErr)
	assert.Equal(t, errcodes.ErrCodeValidation, appErr.ErrorCode)
	require.Len(t, appErr.Fields, 1)
	assert.Equal(t, "otp", appErr.Fields[0].Field)
}

func TestValidationError(t *testing.T) {
	t.Parallel()

	t.Run("returns InvalidArgument for validation errors with field details", func(t *testing.T) {
		t.Parallel()
		v := validator.New()
		type S struct {
			Email string `validate:"required,email"`
			Phone string `validate:"required"`
		}
		validationErr := v.Struct(S{})
		require.Error(t, validationErr)

		result := ValidationError(validationErr, validationDesc)
		st, ok := status.FromError(result)
		require.True(t, ok)
		assert.Equal(t, codes.InvalidArgument, st.Code())

		br := statusFindDetail[*errdetails.BadRequest](st)
		require.NotNil(t, br)
		assert.NotEmpty(t, br.FieldViolations)

		appErr := statusFindDetail[*pb.AppError](st)
		require.NotNil(t, appErr)
		assert.Equal(t, errcodes.ErrCodeValidation, appErr.ErrorCode)
		assert.NotEmpty(t, appErr.Fields)
	})

	t.Run("returns Internal for non-validation error", func(t *testing.T) {
		t.Parallel()
		result := ValidationError(errors.New("not a validation error"), validationDesc)
		st, ok := status.FromError(result)
		require.True(t, ok)
		assert.Equal(t, codes.Internal, st.Code())

		appErr := statusFindDetail[*pb.AppError](st)
		require.NotNil(t, appErr)
		assert.Equal(t, errcodes.ErrCodeInternal, appErr.ErrorCode)
	})
}

func TestCardPreconditionError(t *testing.T) {
	t.Parallel()

	err := CardPreconditionError("card-123", "Card is blocked")
	st, ok := status.FromError(err)
	require.True(t, ok)
	assert.Equal(t, codes.FailedPrecondition, st.Code())

	pf := statusFindDetail[*errdetails.PreconditionFailure](st)
	require.NotNil(t, pf)
	require.Len(t, pf.Violations, 1)
	assert.Equal(t, "Card", pf.Violations[0].Type)
	assert.Equal(t, "card-123", pf.Violations[0].Subject)
	assert.Equal(t, "Card is blocked", pf.Violations[0].Description)
}

func TestPaymentPreconditionError(t *testing.T) {
	t.Parallel()

	tests := []struct {
		condition       payments.RequiredActionType
		expectedSubject string
		expectedDesc    string
	}{
		{payments.RequiredActionTypeIPAddress, "ipAddress", "An ip address is required"},
		{payments.RequiredActionTypeOTP, "otp", "OTP is required"},
		{payments.RequiredActionTypeSenderAmount, "senderAmount", "Amount is required"},
		{payments.RequiredActionTypeSenderAccount, "senderAccount", "Account is required"},
		{payments.RequiredActionTypeSenderIdentifier, "senderIdentifier", "Sender is required"},
		{payments.RequiredActionTypeReceiverAmount, "receiverAmount", "Amount is required"},
		{payments.RequiredActionTypeReceiverIdentifier, "receiverIdentifier", "Recipient is required"},
		{payments.RequiredActionTypeThreeDS, "threeDS", "3DS is required"},
	}

	for _, tc := range tests {
		t.Run(tc.expectedSubject, func(t *testing.T) {
			t.Parallel()
			err := PaymentPreconditionError([]payments.RequiredActionType{tc.condition})
			st, ok := status.FromError(err)
			require.True(t, ok)
			assert.Equal(t, codes.FailedPrecondition, st.Code())

			pf := statusFindDetail[*errdetails.PreconditionFailure](st)
			require.NotNil(t, pf)
			require.Len(t, pf.Violations, 1)
			assert.Equal(t, "Payment", pf.Violations[0].Type)
			assert.Equal(t, tc.expectedSubject, pf.Violations[0].Subject)
			assert.Equal(t, tc.expectedDesc, pf.Violations[0].Description)
		})
	}

	t.Run("unknown condition is skipped", func(t *testing.T) {
		t.Parallel()
		err := PaymentPreconditionError([]payments.RequiredActionType{payments.RequiredActionTypeUnknown})
		st, ok := status.FromError(err)
		require.True(t, ok)
		assert.Equal(t, codes.FailedPrecondition, st.Code())

		pf := statusFindDetail[*errdetails.PreconditionFailure](st)
		require.NotNil(t, pf)
		assert.Empty(t, pf.Violations)
	})

	t.Run("multiple conditions produce multiple violations", func(t *testing.T) {
		t.Parallel()
		err := PaymentPreconditionError([]payments.RequiredActionType{
			payments.RequiredActionTypeOTP,
			payments.RequiredActionTypeIPAddress,
		})
		st, ok := status.FromError(err)
		require.True(t, ok)

		pf := statusFindDetail[*errdetails.PreconditionFailure](st)
		require.NotNil(t, pf)
		assert.Len(t, pf.Violations, 2)
	})
}

func TestPaymentInsufficientFundsError(t *testing.T) {
	t.Parallel()

	err := PaymentInsufficientFundsError()
	st, ok := status.FromError(err)
	require.True(t, ok)
	assert.Equal(t, codes.FailedPrecondition, st.Code())

	pf := statusFindDetail[*errdetails.PreconditionFailure](st)
	require.NotNil(t, pf)
	require.Len(t, pf.Violations, 1)
	assert.Equal(t, "Payment", pf.Violations[0].Type)
	assert.Equal(t, "insufficientFunds", pf.Violations[0].Subject)

	appErr := statusFindDetail[*pb.AppError](st)
	require.NotNil(t, appErr)
	assert.Equal(t, errcodes.ErrCodePaymentsInsufficientFunds, appErr.ErrorCode)
}

func TestSimpleErrorHelpers(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		fn              func(string) error
		msg             string
		expectedCode    codes.Code
		msgContains     string
		expectedAppCode errcodes.AppErrorCode
	}{
		{"InternalError", InternalError, "db failure", codes.Internal, "db failure", errcodes.ErrCodeInternal},
		{"ForbiddenError", ForbiddenError, "read access", codes.PermissionDenied, "read access", errcodes.ErrCodeForbidden},
		{"UnauthenticatedError", UnauthenticatedError, "missing token", codes.Unauthenticated, "missing token", errcodes.ErrCodeUnauthorized},
		{"NotFoundError", NotFoundError, "resource", codes.NotFound, "resource", errcodes.ErrCodeNotFound},
		{"AlreadyExistsError", AlreadyExistsError, "duplicate name", codes.AlreadyExists, "duplicate name", errcodes.ErrCodeConflict},
		{"FailedPreconditionError", FailedPreconditionError, "missing field", codes.FailedPrecondition, "missing field", errcodes.ErrCodeBadRequest},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			err := tc.fn(tc.msg)
			st, ok := status.FromError(err)
			require.True(t, ok)
			assert.Equal(t, tc.expectedCode, st.Code())
			assert.Contains(t, st.Message(), tc.msgContains)

			appErr := statusFindDetail[*pb.AppError](st)
			require.NotNil(t, appErr)
			assert.Equal(t, tc.expectedAppCode, appErr.ErrorCode)
		})
	}
}
