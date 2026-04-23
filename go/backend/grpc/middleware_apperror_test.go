package grpc

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	pb "gitlab.com/fynbos/proto/backend/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func ctxWithRequestId(id string) context.Context {
	return context.WithValue(context.Background(), ctxKeyRequestId, id)
}

func TestWithAppError(t *testing.T) {
	t.Parallel()

	t.Run("raw error becomes Internal with new AppError", func(t *testing.T) {
		t.Parallel()
		err := withAppError(context.Background(), errors.New("db failure"))
		st, ok := status.FromError(err)
		require.True(t, ok)
		assert.Equal(t, codes.Internal, st.Code())
		assert.Equal(t, "db failure", st.Message())

		appErr := statusFindDetail[*pb.AppError](st)
		require.NotNil(t, appErr)
		assert.Equal(t, pb.ErrorCode_ERROR_CODE_INTERNAL.String(), appErr.ErrorCode)
		assert.Equal(t, "db failure", appErr.Message)
	})

	t.Run("grpc status without AppError gets one added", func(t *testing.T) {
		t.Parallel()
		original := status.Error(codes.NotFound, "thing not found")
		err := withAppError(context.Background(), original)
		st, ok := status.FromError(err)
		require.True(t, ok)
		assert.Equal(t, codes.NotFound, st.Code())

		appErr := statusFindDetail[*pb.AppError](st)
		require.NotNil(t, appErr)
		assert.Equal(t, pb.ErrorCode_ERROR_CODE_INTERNAL.String(), appErr.ErrorCode)
		assert.Equal(t, "thing not found", appErr.Message)
	})

	t.Run("grpc status with existing AppError preserves it", func(t *testing.T) {
		t.Parallel()
		base := status.New(codes.PermissionDenied, "forbidden")
		original := statusWithDetails(base, &pb.AppError{
			ErrorCode: pb.ErrorCode_ERROR_CODE_FORBIDDEN.String(),
			Message:   "access denied",
		}).Err()

		err := withAppError(context.Background(), original)
		st, ok := status.FromError(err)
		require.True(t, ok)
		assert.Equal(t, codes.PermissionDenied, st.Code())

		appErr := statusFindDetail[*pb.AppError](st)
		require.NotNil(t, appErr)
		assert.Equal(t, pb.ErrorCode_ERROR_CODE_FORBIDDEN.String(), appErr.ErrorCode)
		assert.Equal(t, "access denied", appErr.Message)
	})

	t.Run("request id from context is set on AppError", func(t *testing.T) {
		t.Parallel()
		ctx := ctxWithRequestId("req-abc-123")
		err := withAppError(ctx, errors.New("some error"))
		st, ok := status.FromError(err)
		require.True(t, ok)

		appErr := statusFindDetail[*pb.AppError](st)
		require.NotNil(t, appErr)
		assert.Equal(t, "req-abc-123", appErr.ReqId)
	})

	t.Run("panics if ReqId is different between ctx and appError", func(t *testing.T) {
		t.Parallel()
		ctx := ctxWithRequestId("ctx-req-id")
		base := status.New(codes.Internal, "error")
		original := statusWithDetails(base, &pb.AppError{
			ErrorCode: pb.ErrorCode_ERROR_CODE_INTERNAL.String(),
			ReqId:     "unexpected-req-id",
		}).Err()

		require.Panics(t, func() {
			withAppError(ctx, original)
		})
	})

	t.Run("empty request id in context sets empty req_id", func(t *testing.T) {
		t.Parallel()
		err := withAppError(context.Background(), errors.New("error"))
		st, ok := status.FromError(err)
		require.True(t, ok)

		appErr := statusFindDetail[*pb.AppError](st)
		require.NotNil(t, appErr)
		assert.Equal(t, "", appErr.ReqId)
	})
}
