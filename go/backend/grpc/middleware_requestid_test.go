package grpc

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/metadata"
)

func TestCtxWithRequestIdFromMeta(t *testing.T) {
	t.Parallel()

	t.Run("sets request id from x-request-id metadata", func(t *testing.T) {
		t.Parallel()
		meta := metadata.Pairs("x-request-id", "test-req-123")
		ctx := ctxWithRequestIdFromMeta(context.Background(), meta)
		assert.Equal(t, "test-req-123", RequestIdFromContext(ctx))
	})

	t.Run("sets empty string when x-request-id is absent", func(t *testing.T) {
		t.Parallel()
		meta := metadata.New(nil)
		ctx := ctxWithRequestIdFromMeta(context.Background(), meta)
		assert.Equal(t, "", RequestIdFromContext(ctx))
	})
}
