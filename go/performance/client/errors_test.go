package client

import (
	"context"
	"errors"
	"testing"

	"github.com/interledger/interledger-app/go/backend/errcodes"
	pb "github.com/interledger/interledger-app/go/proto/backend/v1"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// appErrorStatus builds the shape the backend actually returns: a status whose
// details carry an AppError, as attached by MakeUnaryInterceptorAppError.
func appErrorStatus(t *testing.T, code codes.Code, appErr *pb.AppError) error {
	t.Helper()
	st, err := status.New(code, appErr.GetMessage()).WithDetails(appErr)
	require.NoError(t, err)
	return st.Err()
}

func TestClassifyNilError(t *testing.T) {
	assert.Nil(t, Classify("create", nil))
}

func TestClassifyInsufficientFundsIsExhausted(t *testing.T) {
	// The finish line of a drain run, not a failure.
	err := appErrorStatus(t, codes.FailedPrecondition, &pb.AppError{
		ErrorCode: errcodes.ErrCodePaymentsInsufficientFunds,
		Message:   "insufficient funds",
	})

	f := Classify("confirm", err)
	require.NotNil(t, f)
	assert.Equal(t, ClassExhausted, f.Class)
	assert.Equal(t, errcodes.ErrCodePaymentsInsufficientFunds, f.AppCode)
	assert.Equal(t, errcodes.ErrCodePaymentsInsufficientFunds, f.Key())
}

func TestClassifyKYCLimitIsExhausted(t *testing.T) {
	// Limits arrive as a plain validation error on the amount field, with no
	// dedicated error code, so the message is all there is to match on.
	limits := []string{
		"Exceeds per transaction limit.",
		"Exceeds daily limit.",
		"Exceeds monthly limit.",
		"Exceeds 6 monthly limit.",
		"Exceeds yearly limit.",
		"Exceeds account limit.",
	}

	for _, msg := range limits {
		t.Run(msg, func(t *testing.T) {
			err := appErrorStatus(t, codes.InvalidArgument, &pb.AppError{
				ErrorCode: errcodes.ErrCodeValidation,
				Message:   "validation",
				Fields:    []*pb.AppErrorField{{Field: "amount", Error: msg}},
			})

			f := Classify("create", err)
			require.NotNil(t, f)
			assert.Equal(t, ClassExhausted, f.Class, "a reached KYC limit ends the sender cleanly")
			assert.Contains(t, f.Message, msg)
		})
	}
}

func TestClassifyOtherValidationIsFatal(t *testing.T) {
	// A genuinely malformed request means the scenario is wrong; retrying it a
	// hundred times only produces noise.
	err := appErrorStatus(t, codes.InvalidArgument, &pb.AppError{
		ErrorCode: errcodes.ErrCodeValidation,
		Message:   "validation",
		Fields:    []*pb.AppErrorField{{Field: "amount", Error: "You can't send payments to https://ilp.link/nope"}},
	})

	f := Classify("create", err)
	require.NotNil(t, f)
	assert.Equal(t, ClassFatal, f.Class)
}

func TestClassifySetupErrorsAreFatal(t *testing.T) {
	fatalCodes := []string{
		errcodes.ErrCodeUnauthorized,
		errcodes.ErrCodeForbidden,
		errcodes.ErrCodeUserAAL2Required,
		errcodes.ErrCodeWalletsNoWalletFound,
		errcodes.ErrCodeWalletNotActivated,
		errcodes.ErrCodeLinkedAccNotFound,
		errcodes.ErrCodePaymentsRequiredActions,
	}

	for _, code := range fatalCodes {
		t.Run(code, func(t *testing.T) {
			err := appErrorStatus(t, codes.Unauthenticated, &pb.AppError{ErrorCode: code, Message: code})
			f := Classify("create", err)
			require.NotNil(t, f)
			assert.Equal(t, ClassFatal, f.Class)
		})
	}
}

func TestClassifyUnavailableIsTransient(t *testing.T) {
	// A dropped port-forward or a restarting backend: count it, keep going.
	f := Classify("create", status.Error(codes.Unavailable, "connection refused"))
	require.NotNil(t, f)
	assert.Equal(t, ClassTransient, f.Class)
	assert.Equal(t, "Unavailable", f.Key(), "with no AppError, the gRPC code is the grouping key")
}

func TestClassifyDeadlineExceededIsTransient(t *testing.T) {
	f := Classify("confirm", status.Error(codes.DeadlineExceeded, "context deadline exceeded"))
	require.NotNil(t, f)
	assert.Equal(t, ClassTransient, f.Class)
}

func TestClassifyBareContextDeadline(t *testing.T) {
	// Our own per-request timeout can surface without a gRPC status attached.
	f := Classify("poll", context.DeadlineExceeded)
	require.NotNil(t, f)
	assert.Equal(t, codes.DeadlineExceeded, f.Code)
	assert.Equal(t, ClassTransient, f.Class)
}

func TestClassifyUnauthenticatedWithoutAppErrorIsFatal(t *testing.T) {
	f := Classify("create", status.Error(codes.Unauthenticated, "Unauthenticated"))
	require.NotNil(t, f)
	assert.Equal(t, ClassFatal, f.Class)
}

func TestFailureUnwrapsAndFormats(t *testing.T) {
	underlying := status.Error(codes.Unavailable, "boom")
	f := Classify("create", underlying)
	require.NotNil(t, f)

	assert.ErrorIs(t, f, underlying, "the original error stays reachable for errors.Is")
	assert.Contains(t, f.Error(), "create")
	assert.Contains(t, f.Error(), "Unavailable")
}

func TestClassStrings(t *testing.T) {
	assert.Equal(t, "transient", ClassTransient.String())
	assert.Equal(t, "exhausted", ClassExhausted.String())
	assert.Equal(t, "fatal", ClassFatal.String())
}

func TestClassifyNonStatusError(t *testing.T) {
	f := Classify("create", errors.New("something else"))
	require.NotNil(t, f)
	assert.Equal(t, ClassTransient, f.Class)
	assert.Contains(t, f.Message, "something else")
}
