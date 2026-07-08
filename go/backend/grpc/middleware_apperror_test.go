package grpc

import (
	"context"
	"errors"
	"testing"

	"github.com/interledger/interledger-app/go/backend/appcontext"
	"github.com/interledger/interledger-app/go/backend/errcodes"
	pb "github.com/interledger/interledger-app/go/proto/backend/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func ctxWithRequestId(id string) context.Context {
	return context.WithValue(context.Background(), appcontext.KeyRequestID, id)
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
		assert.Equal(t, errcodes.ErrCodeInternal, appErr.ErrorCode)
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
		assert.Equal(t, errcodes.ErrCodeInternal, appErr.ErrorCode)
		assert.Equal(t, "thing not found", appErr.Message)
	})

	t.Run("grpc status with existing AppError preserves it", func(t *testing.T) {
		t.Parallel()
		base := status.New(codes.PermissionDenied, "forbidden")
		original := statusWithDetails(base, &pb.AppError{
			ErrorCode: errcodes.ErrCodeForbidden,
			Message:   "access denied",
		}).Err()

		err := withAppError(context.Background(), original)
		st, ok := status.FromError(err)
		require.True(t, ok)
		assert.Equal(t, codes.PermissionDenied, st.Code())

		appErr := statusFindDetail[*pb.AppError](st)
		require.NotNil(t, appErr)
		assert.Equal(t, errcodes.ErrCodeForbidden, appErr.ErrorCode)
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
