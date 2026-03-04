package storage

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"gitlab.com/fynbos/mock/mockxago/internal/models"
)

func TestMemoryStorage_SaveAccessToken(t *testing.T) {
	store := NewMemoryStorage()
	token := &models.AccessToken{
		ID:        "test-id",
		Token:     "test-token",
		ExpiresAt: time.Now().Add(1 * time.Hour),
	}

	err := store.SaveAccessToken(context.Background(), token)
	assert.NoError(t, err)

	retrieved, err := store.GetAccessToken(context.Background(), "test-token")
	assert.NoError(t, err)
	assert.NotNil(t, retrieved)
	assert.Equal(t, "test-token", retrieved.Token)
}

func TestMemoryStorage_GetAccessToken_NotFound(t *testing.T) {
	store := NewMemoryStorage()

	_, err := store.GetAccessToken(context.Background(), "non-existent")
	assert.Error(t, err)
	assert.Equal(t, ErrTokenNotFound, err)
}

func TestMemoryStorage_GetAccessToken_Expired(t *testing.T) {
	store := NewMemoryStorage()
	token := &models.AccessToken{
		ID:        "test-id",
		Token:     "expired-token",
		ExpiresAt: time.Now().Add(-1 * time.Hour), // Expired 1 hour ago
	}

	store.SaveAccessToken(context.Background(), token)

	_, err := store.GetAccessToken(context.Background(), "expired-token")
	assert.Error(t, err)
	assert.Equal(t, ErrTokenExpired, err)
}

func TestMemoryStorage_InvalidateAccessToken(t *testing.T) {
	store := NewMemoryStorage()
	token := &models.AccessToken{
		ID:        "test-id",
		Token:     "test-token",
		ExpiresAt: time.Now().Add(1 * time.Hour),
	}

	store.SaveAccessToken(context.Background(), token)
	err := store.InvalidateAccessToken(context.Background(), "test-token")
	assert.NoError(t, err)

	_, err = store.GetAccessToken(context.Background(), "test-token")
	assert.Error(t, err)
}
