package storage

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/interledger/interledger-app/go/mock/mockxago/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Note: Redis tests require a running Redis instance.
// They are skipped in CI if REDIS_TEST_URL is not set.
func redisTestURL() string {
	url := os.Getenv("REDIS_TEST_URL")
	if url == "" {
		url = "redis://localhost:6379"
	}
	return url
}

func skipIfNoRedis(t *testing.T) {
	if os.Getenv("CI") == "" && os.Getenv("REDIS_TEST_ENABLED") == "" {
		t.Skip("Redis tests skipped (set REDIS_TEST_ENABLED=1 to run)")
	}
}

func TestRedisStorage_NewRedisStorage(t *testing.T) {
	skipIfNoRedis(t)

	store, err := NewRedisStorage(redisTestURL(), 1)
	if err != nil {
		t.Skip("Redis not available:", err)
	}
	require.NoError(t, err)
	assert.NotNil(t, store)
}

func TestRedisStorage_SaveAndGetAccessToken(t *testing.T) {
	skipIfNoRedis(t)

	store, err := NewRedisStorage(redisTestURL(), 1)
	if err != nil {
		t.Skip("Redis not available:", err)
	}
	require.NoError(t, err)
	defer store.Reset(context.Background())

	token := &models.AccessToken{
		ID:        "token-123",
		Token:     "test-token-value",
		ExpiresAt: time.Now().Add(1 * time.Hour),
		CreatedAt: time.Now(),
	}

	err = store.SaveAccessToken(context.Background(), token)
	assert.NoError(t, err)

	retrieved, err := store.GetAccessToken(context.Background(), token.Token)
	assert.NoError(t, err)
	assert.NotNil(t, retrieved)
	assert.Equal(t, token.Token, retrieved.Token)
}

func TestRedisStorage_SaveAndGetSubAccount(t *testing.T) {
	skipIfNoRedis(t)

	store, err := NewRedisStorage(redisTestURL(), 1)
	if err != nil {
		t.Skip("Redis not available:", err)
	}
	require.NoError(t, err)
	defer store.Reset(context.Background())

	account := &models.SubAccount{
		ID:        "acc-123",
		WalletID:  "wal-456",
		AccountID: "account-789",
		FirstName: "John",
		LastName:  "Doe",
		Email:     "john@example.com",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err = store.SaveSubAccount(context.Background(), account)
	assert.NoError(t, err)

	retrieved, err := store.GetSubAccount(context.Background(), account.AccountID)
	assert.NoError(t, err)
	assert.NotNil(t, retrieved)
	assert.Equal(t, account.AccountID, retrieved.AccountID)
}

func TestRedisStorage_GetBalance(t *testing.T) {
	skipIfNoRedis(t)

	store, err := NewRedisStorage(redisTestURL(), 1)
	if err != nil {
		t.Skip("Redis not available:", err)
	}
	require.NoError(t, err)
	defer store.Reset(context.Background())

	walletID := "wallet-test"
	currency := "USD"

	// Initially, balance should be 0
	available, reserved, err := store.GetBalance(context.Background(), walletID, currency)
	assert.NoError(t, err)
	assert.Equal(t, 0.0, available)
	assert.Equal(t, 0.0, reserved)
}

func TestRedisStorage_SetBalance(t *testing.T) {
	skipIfNoRedis(t)

	store, err := NewRedisStorage(redisTestURL(), 1)
	if err != nil {
		t.Skip("Redis not available:", err)
	}
	require.NoError(t, err)
	defer store.Reset(context.Background())

	walletID := "wallet-test"
	currency := "USD"

	err = store.SetBalance(context.Background(), walletID, currency, 100.0, 50.0)
	assert.NoError(t, err)

	available, reserved, err := store.GetBalance(context.Background(), walletID, currency)
	assert.NoError(t, err)
	assert.Equal(t, 100.0, available)
	assert.Equal(t, 50.0, reserved)
}

func TestRedisStorage_SaveAndGetTransaction(t *testing.T) {
	skipIfNoRedis(t)

	store, err := NewRedisStorage(redisTestURL(), 1)
	if err != nil {
		t.Skip("Redis not available:", err)
	}
	require.NoError(t, err)
	defer store.Reset(context.Background())

	txn := &models.Transaction{
		ID:        "txn-123",
		AccountID: "acc-456",
		WalletID:  "wal-789",
		Status:    "pending",
		CreatedAt: time.Now(),
	}

	err = store.SaveTransaction(context.Background(), txn)
	assert.NoError(t, err)

	retrieved, err := store.GetTransaction(context.Background(), txn.ID)
	assert.NoError(t, err)
	assert.NotNil(t, retrieved)
	assert.Equal(t, txn.ID, retrieved.ID)
}

func TestRedisStorage_SaveAndGetDeposit(t *testing.T) {
	skipIfNoRedis(t)

	store, err := NewRedisStorage(redisTestURL(), 1)
	if err != nil {
		t.Skip("Redis not available:", err)
	}
	require.NoError(t, err)
	defer store.Reset(context.Background())

	deposit := &models.Deposit{
		ID:               "dep-123",
		AccountID:        "acc-456",
		DepositReference: "ref-abc",
		Status:           "pending",
		CreatedAt:        time.Now(),
	}

	err = store.SaveDeposit(context.Background(), deposit)
	assert.NoError(t, err)

	retrieved, err := store.GetDeposit(context.Background(), deposit.ID)
	assert.NoError(t, err)
	assert.NotNil(t, retrieved)
	assert.Equal(t, deposit.ID, retrieved.ID)
}

func TestRedisStorage_SaveAndGetBeneficiary(t *testing.T) {
	skipIfNoRedis(t)

	store, err := NewRedisStorage(redisTestURL(), 1)
	if err != nil {
		t.Skip("Redis not available:", err)
	}
	require.NoError(t, err)
	defer store.Reset(context.Background())

	ben := &models.Beneficiary{
		ID:        "ben-123",
		AccountID: "acc-456",
		WalletID:  "wal-789",
		Name:      "John Beneficiary",
		Status:    "active",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err = store.SaveBeneficiary(context.Background(), ben)
	assert.NoError(t, err)

	retrieved, err := store.GetBeneficiary(context.Background(), ben.ID)
	assert.NoError(t, err)
	assert.NotNil(t, retrieved)
	assert.Equal(t, ben.ID, retrieved.ID)
}
