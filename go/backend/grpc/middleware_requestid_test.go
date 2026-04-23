package grpc

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/metadata"
)

func TestCtxWithRequestIdFromMeta(t *testing.T) {
	t.Parallel()

	t.Run("sets request id from metadata", func(t *testing.T) {
		t.Parallel()
		meta := metadata.Pairs(metaRequestIdKey, "test-req-123")
		ctx, err := ctxWithRequestIdFromMeta(context.Background(), meta)
		assert.NoError(t, err)
		assert.Equal(t, "test-req-123", RequestIdFromContext(ctx))
	})

	t.Run("sets empty string when request id is absent", func(t *testing.T) {
		t.Parallel()
		meta := metadata.New(nil)
		ctx, err := ctxWithRequestIdFromMeta(context.Background(), meta)
		assert.NoError(t, err)
		assert.Equal(t, "", RequestIdFromContext(ctx))
	})

	t.Run("validates special chars in request id", func(t *testing.T) {
		t.Parallel()
		meta := metadata.Pairs(metaRequestIdKey, "invalid request id")
		_, err := ctxWithRequestIdFromMeta(context.Background(), meta)
		assert.Error(t, err)
	})

	t.Run("validates request id length", func(t *testing.T) {
		t.Parallel()
		longRequestId := strings.Repeat("a", requestIdValidationMaxLen+1)
		meta := metadata.Pairs(metaRequestIdKey, longRequestId)
		_, err := ctxWithRequestIdFromMeta(context.Background(), meta)
		assert.Error(t, err)
	})
}
