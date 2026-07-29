package client

import (
	"context"
	"errors"
	"strings"

	"github.com/interledger/interledger-app/go/backend/errcodes"
	pb "github.com/interledger/interledger-app/go/proto/backend/v1"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Class describes how the runner should react to a failed RPC.
type Class int

const (
	// ClassTransient is a failure worth counting but not worth stopping for:
	// timeouts, Unavailable, and anything else that may succeed next time.
	ClassTransient Class = iota
	// ClassExhausted means this sender cannot usefully continue — its balance is
	// gone or it has hit a KYC limit. The sender stops; the run continues.
	ClassExhausted
	// ClassFatal means the harness is misconfigured (bad credentials, unknown
	// receiver, unsupported payment). Retrying would only produce noise, so the
	// whole run stops.
	ClassFatal
)

func (c Class) String() string {
	switch c {
	case ClassExhausted:
		return "exhausted"
	case ClassFatal:
		return "fatal"
	default:
		return "transient"
	}
}

// Failure is a classified RPC error, carrying enough detail for the report to
// explain what the backend actually said.
type Failure struct {
	// Stage is the RPC that failed, e.g. "create" or "confirm".
	Stage string
	// Code is the gRPC status code.
	Code codes.Code
	// AppCode is the backend's own error code from the AppError status detail,
	// which is far more specific than the gRPC code. Empty when absent.
	AppCode string
	// Message is the human-readable reason.
	Message string
	// Class is how the runner should react.
	Class Class

	err error
}

func (f *Failure) Error() string {
	if f.AppCode != "" {
		return f.Stage + ": " + f.AppCode + ": " + f.Message
	}
	return f.Stage + ": " + f.Code.String() + ": " + f.Message
}

func (f *Failure) Unwrap() error { return f.err }

// Key is a stable label for grouping failures in metrics and in the report.
func (f *Failure) Key() string {
	if f.AppCode != "" {
		return f.AppCode
	}
	return f.Code.String()
}

// Classify turns a gRPC error into a Failure.
//
// Every backend handler funnels its errors through toGRPCError, which attaches an
// AppError detail alongside the status. Reading that detail is what lets the
// harness distinguish "this wallet is empty" (expected, and the point of a drain
// run) from "the system is failing" (the thing we are measuring for).
func Classify(stage string, err error) *Failure {
	if err == nil {
		return nil
	}

	f := &Failure{Stage: stage, err: err, Code: codes.Unknown, Message: err.Error()}

	st, ok := status.FromError(err)
	if !ok {
		// Not a status error: a context deadline from our own per-request timeout,
		// most likely.
		if errors.Is(err, context.DeadlineExceeded) {
			f.Code = codes.DeadlineExceeded
		}
		f.Class = ClassTransient
		return f
	}

	f.Code = st.Code()
	f.Message = st.Message()

	for _, detail := range st.Details() {
		appErr, ok := detail.(*pb.AppError)
		if !ok {
			continue
		}
		f.AppCode = appErr.GetErrorCode()
		if msg := appErr.GetMessage(); msg != "" {
			f.Message = msg
		}
		// Validation errors carry the offending field, which is where the KYC
		// limit messages land.
		for _, field := range appErr.GetFields() {
			if field.GetError() != "" {
				f.Message = field.GetField() + ": " + field.GetError()
			}
		}
		break
	}

	f.Class = classify(f)
	return f
}

func classify(f *Failure) Class {
	switch f.AppCode {
	case errcodes.ErrCodePaymentsInsufficientFunds:
		// The wallet is empty. In a drain run this is the finish line.
		return ClassExhausted

	case errcodes.ErrCodeUnauthorized,
		errcodes.ErrCodeForbidden,
		errcodes.ErrCodeUserNoUserFound,
		errcodes.ErrCodeUserAAL1Required,
		errcodes.ErrCodeUserAAL2Required,
		errcodes.ErrCodeWalletsNoWalletFound,
		errcodes.ErrCodeWalletNotActivated,
		errcodes.ErrCodeLinkedAccNotFound,
		errcodes.ErrCodeKYCResubmissionRequired,
		errcodes.ErrCodePaymentsRequiredActions:
		// Setup problems and payments that need a human (3DS). No amount of
		// retrying fixes these, and a run full of them measures nothing.
		return ClassFatal

	case errcodes.ErrCodeValidation:
		// Validation covers both KYC transaction/daily/monthly limits and genuinely
		// malformed requests. Limits are the common case and read as exhaustion
		// rather than as a broken harness — see isLimitMessage.
		if isLimitMessage(f.Message) {
			return ClassExhausted
		}
		return ClassFatal
	}

	switch f.Code {
	case codes.Unauthenticated, codes.PermissionDenied, codes.InvalidArgument, codes.NotFound:
		return ClassFatal
	case codes.ResourceExhausted:
		return ClassExhausted
	default:
		return ClassTransient
	}
}

// isLimitMessage recognises the KYC limit rejections raised by CreatePayment,
// UpdatePayment and ConfirmPayment. They arrive as plain validation errors on the
// "amount" field with no dedicated error code, so the message is all there is to
// match on.
func isLimitMessage(msg string) bool {
	return strings.Contains(msg, "Exceeds ") && strings.Contains(msg, " limit.")
}
