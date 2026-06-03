package storage

import (
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"gitlab.com/fynbos/mock/mockgatehub/internal/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestRedisStore(t *testing.T) (*RedisStorage, *miniredis.Miniredis) {
	t.Helper()
	mini := miniredis.RunT(t)
	store, err := NewRedisStorage("redis://"+mini.Addr(), 0)
	require.NoError(t, err)
	t.Cleanup(func() { store.Close() })
	return store, mini
}

func TestRedisStorage(t *testing.T) {
	store, _ := newTestRedisStore(t)

	t.Run("User Operations", func(t *testing.T) {
		user := &models.User{Email: "redis@test.com"}
		err := store.CreateUser(user)
		require.NoError(t, err)
		assert.NotEmpty(t, user.ID)

		retrieved, err := store.GetUser(user.ID)
		require.NoError(t, err)
		assert.Equal(t, user.Email, retrieved.Email)

		retrievedByEmail, err := store.GetUserByEmail(user.Email)
		require.NoError(t, err)
		assert.Equal(t, user.ID, retrievedByEmail.ID)

		user.Email = "updated@test.com"
		err = store.UpdateUser(user)
		require.NoError(t, err)

		updated, err := store.GetUser(user.ID)
		require.NoError(t, err)
		assert.Equal(t, "updated@test.com", updated.Email)
	})

	t.Run("Wallet Operations", func(t *testing.T) {
		user := &models.User{Email: "wallet-user@test.com"}
		err := store.CreateUser(user)
		require.NoError(t, err)

		wallet := &models.Wallet{
			Address: "rTestAddress123",
			UserID:  user.ID,
			Name:    "Test Wallet",
		}

		err = store.CreateWallet(wallet)
		require.NoError(t, err)
		assert.NotZero(t, wallet.CreatedAt)

		retrieved, err := store.GetWallet(wallet.Address)
		require.NoError(t, err)
		assert.Equal(t, wallet.UserID, retrieved.UserID)

		wallets, err := store.GetWalletsByUser(user.ID)
		require.NoError(t, err)
		assert.Len(t, wallets, 1)
	})

	t.Run("Transaction Operations", func(t *testing.T) {
		user := &models.User{Email: "tx-user@test.com"}
		err := store.CreateUser(user)
		require.NoError(t, err)

		tx := &models.Transaction{
			UserID:      user.ID,
			Amount:      "100.50",
			TotalAmount: "100.50",
			Fee:         "0.00",
			Currency:    "USD",
			Status:      1, // 1 = completed
		}

		err = store.CreateTransaction(tx)
		require.NoError(t, err)
		assert.NotEmpty(t, tx.ID)
		assert.NotZero(t, tx.CreatedAt)

		retrieved, err := store.GetTransaction(tx.ID)
		require.NoError(t, err)
		assert.Equal(t, tx.Amount, retrieved.Amount)
		assert.Equal(t, "100.50", retrieved.Amount)
		assert.Equal(t, 1, retrieved.Status)
	})

	t.Run("Balance Operations", func(t *testing.T) {
		user := &models.User{Email: "balance-user@test.com"}
		err := store.CreateUser(user)
		require.NoError(t, err)

		balance, err := store.GetBalance(user.ID, "USD")
		require.NoError(t, err)
		assert.Equal(t, 0.0, balance)

		err = store.AddBalance(user.ID, "USD", 100.0)
		require.NoError(t, err)

		balance, err = store.GetBalance(user.ID, "USD")
		require.NoError(t, err)
		assert.Equal(t, 100.0, balance)

		err = store.DeductBalance(user.ID, "USD", 30.0)
		require.NoError(t, err)

		balance, err = store.GetBalance(user.ID, "USD")
		require.NoError(t, err)
		assert.Equal(t, 70.0, balance)
	})
}

// ── ListUsers ──────────────────────────────────────────────────────────────

func TestRedisListUsers_Empty(t *testing.T) {
	store, err := NewRedisStorage("redis://localhost:6379", 15)
	if err != nil {
		t.Skip("Redis not available")
	}
	defer store.Close()
	_ = store.client.FlushDB(store.ctx).Err()

	users, err := store.ListUsers()
	require.NoError(t, err)
	assert.Empty(t, users)
}

func TestRedisListUsers_WithSeededUsers(t *testing.T) {
	store, err := NewRedisStorage("redis://localhost:6379", 15)
	if err != nil {
		t.Skip("Redis not available")
	}
	defer store.Close()
	_ = store.client.FlushDB(store.ctx).Err()

	require.NoError(t, SeedTestUsers(store))

	users, err := store.ListUsers()
	require.NoError(t, err)
	assert.Len(t, users, 2)

	emails := make([]string, 0, len(users))
	for _, u := range users {
		emails = append(emails, u.Email)
	}
	assert.Contains(t, emails, "testuser1@mockgatehub.local")
	assert.Contains(t, emails, "testuser2@mockgatehub.local")
}

// ── ListTransactionsByUser ─────────────────────────────────────────────────

func TestRedisListTransactionsByUser_Empty(t *testing.T) {
	store, err := NewRedisStorage("redis://localhost:6379", 15)
	if err != nil {
		t.Skip("Redis not available")
	}
	defer store.Close()
	_ = store.client.FlushDB(store.ctx).Err()

	txns, err := store.ListTransactionsByUser("some-user-id")
	require.NoError(t, err)
	assert.Empty(t, txns)
}

func TestRedisListTransactionsByUser(t *testing.T) {
	store, err := NewRedisStorage("redis://localhost:6379", 15)
	if err != nil {
		t.Skip("Redis not available")
	}
	defer store.Close()
	_ = store.client.FlushDB(store.ctx).Err()

	user1 := "00000000-0000-0000-0000-000000000001"
	user2 := "00000000-0000-0000-0000-000000000002"

	require.NoError(t, store.CreateTransaction(&models.Transaction{ID: "tx-r1", UserID: user1, Amount: "10.00", Currency: "USD"}))
	require.NoError(t, store.CreateTransaction(&models.Transaction{ID: "tx-r2", UserID: user1, Amount: "20.00", Currency: "EUR"}))
	require.NoError(t, store.CreateTransaction(&models.Transaction{ID: "tx-r3", UserID: user2, Amount: "5.00", Currency: "GBP"}))

	u1Txns, err := store.ListTransactionsByUser(user1)
	require.NoError(t, err)
	assert.Len(t, u1Txns, 2)

	u2Txns, err := store.ListTransactionsByUser(user2)
	require.NoError(t, err)
	assert.Len(t, u2Txns, 1)
	assert.Equal(t, "tx-r3", u2Txns[0].ID)
}

func TestRedisCreateTransaction_AddsUserIndex(t *testing.T) {
	store, err := NewRedisStorage("redis://localhost:6379", 15)
	if err != nil {
		t.Skip("Redis not available")
	}
	defer store.Close()
	_ = store.client.FlushDB(store.ctx).Err()

	userID := "index-test-user"
	require.NoError(t, store.CreateTransaction(&models.Transaction{ID: "idx-tx-1", UserID: userID, Amount: "1.00", Currency: "USD"}))

	ids, err := store.client.LRange(store.ctx, store.userTransactionsKey(userID), 0, -1).Result()
	require.NoError(t, err)
	assert.Contains(t, ids, "idx-tx-1")
}

// ── GetAllBalances ─────────────────────────────────────────────────────────

func TestRedisGetAllBalances_Empty(t *testing.T) {
	store, err := NewRedisStorage("redis://localhost:6379", 15)
	if err != nil {
		t.Skip("Redis not available")
	}
	defer store.Close()
	_ = store.client.FlushDB(store.ctx).Err()

	balances, err := store.GetAllBalances("nonexistent-user")
	require.NoError(t, err)
	assert.NotNil(t, balances)
	assert.Empty(t, balances)
}

func TestRedisGetAllBalances(t *testing.T) {
	store, err := NewRedisStorage("redis://localhost:6379", 15)
	if err != nil {
		t.Skip("Redis not available")
	}
	defer store.Close()
	_ = store.client.FlushDB(store.ctx).Err()

	userID := "balance-list-user"
	require.NoError(t, store.AddBalance(userID, "USD", 100.0))
	require.NoError(t, store.AddBalance(userID, "EUR", 200.0))

	balances, err := store.GetAllBalances(userID)
	require.NoError(t, err)
	assert.Equal(t, 100.0, balances["USD"])
	assert.Equal(t, 200.0, balances["EUR"])
	assert.Len(t, balances, 2)
}

func TestRedisConnectionError(t *testing.T) {
	_, err := NewRedisStorage("redis://localhost:99999", 0)
	assert.Error(t, err)
}

func TestRedisInvalidURL(t *testing.T) {
	_, err := NewRedisStorage("invalid-url", 0)
	assert.Error(t, err)
}

func TestRedisConcurrency(t *testing.T) {
	store, _ := newTestRedisStore(t)

	user := &models.User{Email: "concurrent@test.com"}
	require.NoError(t, store.CreateUser(user))

	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func() {
			for j := 0; j < 100; j++ {
				_ = store.AddBalance(user.ID, "USD", 1.0)
			}
			done <- true
		}()
	}

	for i := 0; i < 10; i++ {
		<-done
	}

	// Allow in-flight increments to settle
	time.Sleep(100 * time.Millisecond)

	balance, err := store.GetBalance(user.ID, "USD")
	require.NoError(t, err)
	assert.Equal(t, 1000.0, balance)
}
