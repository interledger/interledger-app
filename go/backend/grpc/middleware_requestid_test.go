package grpc

import (
	"context"
	"testing"

	"github.com/interledger/interledger-app/go/backend/appcontext"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/metadata"
)

func TestCtxWithRequestIdFromMeta(t *testing.T) {
	t.Parallel()

	t.Run("sets request id from metadata", func(t *testing.T) {
		t.Parallel()
		meta := metadata.Pairs(metaRequestIDKey, "test-req-123")
		ctx, err := ctxWithRequestIdFromMeta(context.Background(), meta)
		assert.NoError(t, err)
		assert.Equal(t, "test-req-123", appcontext.RequestIDFromContext(ctx))
	})

	t.Run("sets empty string when request id is absent", func(t *testing.T) {
		t.Parallel()
		meta := metadata.New(nil)
		ctx, err := ctxWithRequestIdFromMeta(context.Background(), meta)
		assert.NoError(t, err)
		assert.Equal(t, "", appcontext.RequestIDFromContext(ctx))
	})
}
