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

func TestMemoryStorage_SaveSubAccount(t *testing.T) {
	store := NewMemoryStorage()
	account := &models.SubAccount{
		ID:        "sub-id",
		WalletID:  "wallet-1",
		AccountID: "account-1",
		FirstName: "John",
		LastName:  "Doe",
		Email:     "john@example.com",
	}

	err := store.SaveSubAccount(context.Background(), account)
	assert.NoError(t, err)

	retrieved, err := store.GetSubAccount(context.Background(), "account-1")
	assert.NoError(t, err)
	assert.NotNil(t, retrieved)
	assert.Equal(t, "account-1", retrieved.AccountID)
	assert.Equal(t, "John", retrieved.FirstName)
}

func TestMemoryStorage_GetSubAccount_NotFound(t *testing.T) {
	store := NewMemoryStorage()

	_, err := store.GetSubAccount(context.Background(), "non-existent")
	assert.Error(t, err)
	assert.Equal(t, ErrSubAccountNotFound, err)
}

func TestMemoryStorage_GetSubAccountByWalletID(t *testing.T) {
	store := NewMemoryStorage()
	account := &models.SubAccount{
		ID:        "sub-id",
		WalletID:  "wallet-1",
		AccountID: "account-1",
		FirstName: "Jane",
		LastName:  "Smith",
		Email:     "jane@example.com",
	}

	err := store.SaveSubAccount(context.Background(), account)
	assert.NoError(t, err)

	retrieved, err := store.GetSubAccountByWalletID(context.Background(), "wallet-1")
	assert.NoError(t, err)
	assert.NotNil(t, retrieved)
	assert.Equal(t, "account-1", retrieved.AccountID)
	assert.Equal(t, "wallet-1", retrieved.WalletID)
}

func TestMemoryStorage_GetSubAccountByWalletID_NotFound(t *testing.T) {
	store := NewMemoryStorage()

	_, err := store.GetSubAccountByWalletID(context.Background(), "non-existent")
	assert.Error(t, err)
	assert.Equal(t, ErrSubAccountNotFound, err)
}

func TestMemoryStorage_UpdateSubAccount(t *testing.T) {
	store := NewMemoryStorage()
	account := &models.SubAccount{
		ID:              "sub-id",
		WalletID:        "wallet-1",
		AccountID:       "account-1",
		FirstName:       "John",
		LastName:        "Doe",
		Email:           "john@example.com",
		PhysicalAddress: "123 Main St",
	}

	store.SaveSubAccount(context.Background(), account)

	account.PhysicalAddress = "456 Oak Ave"
	err := store.UpdateSubAccount(context.Background(), account)
	assert.NoError(t, err)

	retrieved, err := store.GetSubAccount(context.Background(), "account-1")
	assert.NoError(t, err)
	assert.Equal(t, "456 Oak Ave", retrieved.PhysicalAddress)
}
