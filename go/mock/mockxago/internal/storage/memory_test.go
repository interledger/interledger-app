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

func TestMemoryStorage_GetBalance_NoBalance(t *testing.T) {
	store := NewMemoryStorage()

	available, reserved, err := store.GetBalance(context.Background(), "wallet1", "USD")
	assert.NoError(t, err)
	assert.Equal(t, 0.0, available)
	assert.Equal(t, 0.0, reserved)
}

func TestMemoryStorage_SetBalance(t *testing.T) {
	store := NewMemoryStorage()

	err := store.SetBalance(context.Background(), "wallet1", "USD", 100.0, 10.0)
	assert.NoError(t, err)

	available, reserved, _ := store.GetBalance(context.Background(), "wallet1", "USD")
	assert.Equal(t, 100.0, available)
	assert.Equal(t, 10.0, reserved)
}

func TestMemoryStorage_AddBalance(t *testing.T) {
	store := NewMemoryStorage()

	store.SetBalance(context.Background(), "wallet1", "USD", 100.0, 5.0)
	err := store.AddBalance(context.Background(), "wallet1", "USD", 50.0)
	assert.NoError(t, err)

	available, reserved, _ := store.GetBalance(context.Background(), "wallet1", "USD")
	assert.Equal(t, 150.0, available)
	assert.Equal(t, 5.0, reserved)
}

func TestMemoryStorage_SubtractBalance(t *testing.T) {
	store := NewMemoryStorage()

	store.SetBalance(context.Background(), "wallet1", "USD", 100.0, 0)
	err := store.SubtractBalance(context.Background(), "wallet1", "USD", 30.0)
	assert.NoError(t, err)

	available, _, _ := store.GetBalance(context.Background(), "wallet1", "USD")
	assert.Equal(t, 70.0, available)
}

func TestMemoryStorage_SubtractBalance_InsufficientFunds(t *testing.T) {
	store := NewMemoryStorage()

	store.SetBalance(context.Background(), "wallet1", "USD", 50.0, 0)
	err := store.SubtractBalance(context.Background(), "wallet1", "USD", 100.0)
	assert.Error(t, err)
	assert.Equal(t, ErrInsufficientBalance, err)
}

func TestMemoryStorage_SaveSubAccount(t *testing.T) {
	store := NewMemoryStorage()
	account := &models.SubAccount{
		ID:        "sub1",
		WalletID:  "wallet1",
		AccountID: "acc1",
	}

	err := store.SaveSubAccount(context.Background(), account)
	assert.NoError(t, err)

	retrieved, err := store.GetSubAccount(context.Background(), "acc1")
	assert.NoError(t, err)
	assert.NotNil(t, retrieved)
}

func TestMemoryStorage_GetSubAccountByWalletID(t *testing.T) {
	store := NewMemoryStorage()
	account := &models.SubAccount{
		ID:        "sub1",
		WalletID:  "wallet1",
		AccountID: "acc1",
	}

	store.SaveSubAccount(context.Background(), account)

	retrieved, err := store.GetSubAccountByWalletID(context.Background(), "wallet1")
	assert.NoError(t, err)
	assert.NotNil(t, retrieved)
	assert.Equal(t, "wallet1", retrieved.WalletID)
}

func TestMemoryStorage_SaveBeneficiary(t *testing.T) {
	store := NewMemoryStorage()
	beneficiary := &models.Beneficiary{
		ID:       "ben1",
		WalletID: "wallet1",
		BankName: "Test Bank",
		Status:   "active",
	}

	err := store.SaveBeneficiary(context.Background(), beneficiary)
	assert.NoError(t, err)

	retrieved, err := store.GetBeneficiary(context.Background(), "ben1")
	assert.NoError(t, err)
	assert.NotNil(t, retrieved)
}

func TestMemoryStorage_ListBeneficiariesByAccountID(t *testing.T) {
	store := NewMemoryStorage()

	for i := 0; i < 5; i++ {
		beneficiary := &models.Beneficiary{
			ID:        "ben" + string(rune(i)),
			AccountID: "account1",
			WalletID:  "wallet1",
			BankName:  "Bank " + string(rune(i)),
		}
		store.SaveBeneficiary(context.Background(), beneficiary)
	}

	beneficiaries, total, err := store.ListBeneficiariesByAccountID(context.Background(), "account1", 3, 0)
	assert.NoError(t, err)
	assert.Equal(t, 5, total)
	assert.Equal(t, 3, len(beneficiaries))
}

func TestMemoryStorage_SaveTransaction(t *testing.T) {
	store := NewMemoryStorage()
	tx := &models.Transaction{
		ID:       "tx1",
		WalletID: "wallet1",
		Amount:   100.0,
		Currency: "USD",
		Status:   "pending",
	}

	err := store.SaveTransaction(context.Background(), tx)
	assert.NoError(t, err)

	retrieved, err := store.GetTransaction(context.Background(), "tx1")
	assert.NoError(t, err)
	assert.NotNil(t, retrieved)
}

func TestMemoryStorage_SaveDeposit(t *testing.T) {
	store := NewMemoryStorage()
	deposit := &models.Deposit{
		ID:               "dep1",
		AccountID:        "acc1",
		Amount:           100.0,
		Currency:         "USD",
		DepositReference: "ref123",
		Status:           "pending",
	}

	err := store.SaveDeposit(context.Background(), deposit)
	assert.NoError(t, err)

	retrieved, err := store.GetDeposit(context.Background(), "dep1")
	assert.NoError(t, err)
	assert.NotNil(t, retrieved)
}

func TestMemoryStorage_GetDepositByReference(t *testing.T) {
	store := NewMemoryStorage()
	deposit := &models.Deposit{
		ID:               "dep1",
		AccountID:        "acc1",
		Amount:           100.0,
		Currency:         "USD",
		DepositReference: "ref123",
		Status:           "pending",
	}

	store.SaveDeposit(context.Background(), deposit)

	retrieved, err := store.GetDepositByReference(context.Background(), "ref123")
	assert.NoError(t, err)
	assert.NotNil(t, retrieved)
	assert.Equal(t, "ref123", retrieved.DepositReference)
}

func TestMemoryStorage_SaveTokenAccount(t *testing.T) {
	store := NewMemoryStorage()

	err := store.SaveTokenAccount(context.Background(), "test-token", "acc-123")
	assert.NoError(t, err)

	accountID, err := store.GetAccountIDByToken(context.Background(), "test-token")
	assert.NoError(t, err)
	assert.Equal(t, "acc-123", accountID)
}

func TestMemoryStorage_GetAccountIDByToken_NotFound(t *testing.T) {
	store := NewMemoryStorage()

	_, err := store.GetAccountIDByToken(context.Background(), "non-existent")
	assert.Error(t, err)
}

func TestMemoryStorage_SaveJob(t *testing.T) {
	store := NewMemoryStorage()
	job := &models.Job{
		ID:      "job-1",
		JobType: "webhook",
		Status:  "pending",
	}

	err := store.SaveJob(context.Background(), job)
	assert.NoError(t, err)

	retrieved, err := store.GetJob(context.Background(), "job-1")
	assert.NoError(t, err)
	assert.NotNil(t, retrieved)
	assert.Equal(t, "job-1", retrieved.ID)
}

func TestMemoryStorage_ListReadyJobs(t *testing.T) {
	store := NewMemoryStorage()

	// Create jobs with different NotBefore times
	now := time.Now()
	jobs := []*models.Job{
		{ID: "job-1", JobType: "webhook", Status: "pending", NotBefore: now.Add(-1 * time.Hour)},
		{ID: "job-2", JobType: "webhook", Status: "pending", NotBefore: now.Add(-30 * time.Minute)},
		{ID: "job-3", JobType: "webhook", Status: "pending", NotBefore: now.Add(1 * time.Hour)}, // Not ready yet
	}

	for _, job := range jobs {
		store.SaveJob(context.Background(), job)
	}

	ready, err := store.ListReadyJobs(context.Background(), 10)
	assert.NoError(t, err)
	assert.Len(t, ready, 2) // Only job-1 and job-2 should be ready
}

func TestMemoryStorage_UpdateJobStatus(t *testing.T) {
	store := NewMemoryStorage()
	job := &models.Job{
		ID:      "job-1",
		JobType: "webhook",
		Status:  "pending",
	}

	store.SaveJob(context.Background(), job)

	completedAt := time.Now()
	err := store.UpdateJobStatus(context.Background(), "job-1", "completed", &completedAt, "")
	assert.NoError(t, err)

	updated, _ := store.GetJob(context.Background(), "job-1")
	assert.Equal(t, "completed", updated.Status)
}

func TestMemoryStorage_IncrementJobAttempts(t *testing.T) {
	store := NewMemoryStorage()
	job := &models.Job{
		ID:       "job-1",
		JobType:  "webhook",
		Status:   "pending",
		Attempts: 0,
	}

	store.SaveJob(context.Background(), job)

	err := store.IncrementJobAttempts(context.Background(), "job-1")
	assert.NoError(t, err)

	updated, _ := store.GetJob(context.Background(), "job-1")
	assert.Equal(t, 1, updated.Attempts)
}

func TestMemoryStorage_ClearJobs(t *testing.T) {
	store := NewMemoryStorage()
	job := &models.Job{
		ID:      "job-1",
		JobType: "webhook",
		Status:  "pending",
	}

	store.SaveJob(context.Background(), job)

	err := store.ClearJobs(context.Background())
	assert.NoError(t, err)

	// After clearing, getting a job should return nil
	retrieved, err := store.GetJob(context.Background(), "job-1")
	assert.NoError(t, err)
	assert.Nil(t, retrieved)
}

func TestMemoryStorage_SaveIdempotencyKey(t *testing.T) {
	store := NewMemoryStorage()

	// First create a transaction
	tx := &models.Transaction{
		ID:        "tx-456",
		AccountID: "acc-1",
		WalletID:  "wallet-1",
		Amount:    100.0,
		Currency:  "USD",
		Status:    "pending",
	}
	store.SaveTransaction(context.Background(), tx)

	// Now save the idempotency key
	err := store.SaveIdempotencyKey(context.Background(), "idem-key-123", "tx-456")
	assert.NoError(t, err)

	// Retrieve by idempotency key
	retrieved, err := store.GetTransactionByIdempotencyKey(context.Background(), "idem-key-123")
	assert.NoError(t, err)
	assert.NotNil(t, retrieved)
	assert.Equal(t, "tx-456", retrieved.ID)
}

func TestMemoryStorage_ListDeposits(t *testing.T) {
	store := NewMemoryStorage()

	for i := 0; i < 15; i++ {
		deposit := &models.Deposit{
			ID:        "dep-" + string(rune(i+48)),
			AccountID: "acc-1",
			Amount:    100.0,
			Currency:  "USD",
			Status:    "completed",
		}
		store.SaveDeposit(context.Background(), deposit)
	}

	deposits, total, err := store.ListDeposits(context.Background(), 10, 0)
	assert.NoError(t, err)
	assert.Equal(t, 15, total)
	assert.Len(t, deposits, 10)
}

func TestMemoryStorage_UpdateDepositStatus(t *testing.T) {
	store := NewMemoryStorage()
	deposit := &models.Deposit{
		ID:        "dep-1",
		AccountID: "acc-1",
		Amount:    100.0,
		Currency:  "USD",
		Status:    "pending",
	}

	store.SaveDeposit(context.Background(), deposit)

	err := store.UpdateDepositStatus(context.Background(), "dep-1", "completed")
	assert.NoError(t, err)

	updated, _ := store.GetDeposit(context.Background(), "dep-1")
	assert.Equal(t, "completed", updated.Status)
}

func TestMemoryStorage_UpdateBeneficiaryStatus(t *testing.T) {
	store := NewMemoryStorage()
	beneficiary := &models.Beneficiary{
		ID:       "ben-1",
		WalletID: "wallet-1",
		Status:   "pending",
	}

	store.SaveBeneficiary(context.Background(), beneficiary)

	err := store.UpdateBeneficiaryStatus(context.Background(), "ben-1", "approved")
	assert.NoError(t, err)

	updated, _ := store.GetBeneficiary(context.Background(), "ben-1")
	assert.Equal(t, "approved", updated.Status)
}

func TestMemoryStorage_ListTransactionsByAccount(t *testing.T) {
	store := NewMemoryStorage()

	for i := 0; i < 5; i++ {
		tx := &models.Transaction{
			ID:        "tx-" + string(rune(i+48)),
			AccountID: "acc-1",
			WalletID:  "wallet-1",
			Amount:    100.0,
			Currency:  "USD",
			Status:    "completed",
		}
		store.SaveTransaction(context.Background(), tx)
	}

	txs, total, err := store.ListTransactionsByAccount(context.Background(), "acc-1", 10, 0)
	assert.NoError(t, err)
	assert.Equal(t, 5, total)
	assert.Len(t, txs, 5)
}

func TestMemoryStorage_UpdateTransactionStatus(t *testing.T) {
	store := NewMemoryStorage()
	tx := &models.Transaction{
		ID:        "tx-1",
		AccountID: "acc-1",
		WalletID:  "wallet-1",
		Amount:    100.0,
		Currency:  "USD",
		Status:    "pending",
	}

	store.SaveTransaction(context.Background(), tx)

	err := store.UpdateTransactionStatus(context.Background(), "tx-1", "completed")
	assert.NoError(t, err)

	updated, _ := store.GetTransaction(context.Background(), "tx-1")
	assert.Equal(t, "completed", updated.Status)
}

func TestMemoryStorage_ClearTransactions(t *testing.T) {
	store := NewMemoryStorage()
	tx := &models.Transaction{
		ID:        "tx-1",
		AccountID: "acc-1",
		WalletID:  "wallet-1",
		Amount:    100.0,
		Currency:  "USD",
		Status:    "pending",
	}

	store.SaveTransaction(context.Background(), tx)

	err := store.ClearTransactions(context.Background())
	assert.NoError(t, err)

	_, err = store.GetTransaction(context.Background(), "tx-1")
	assert.Error(t, err)
}

func TestMemoryStorage_ClearDeposits(t *testing.T) {
	store := NewMemoryStorage()
	deposit := &models.Deposit{
		ID:        "dep-1",
		AccountID: "acc-1",
		Amount:    100.0,
		Currency:  "USD",
		Status:    "pending",
	}

	store.SaveDeposit(context.Background(), deposit)

	err := store.ClearDeposits(context.Background())
	assert.NoError(t, err)

	_, err = store.GetDeposit(context.Background(), "dep-1")
	assert.Error(t, err)
}

func TestMemoryStorage_ClearBalances(t *testing.T) {
	store := NewMemoryStorage()

	store.SetBalance(context.Background(), "wallet-1", "USD", 100.0, 10.0)

	err := store.ClearBalances(context.Background())
	assert.NoError(t, err)

	available, reserved, _ := store.GetBalance(context.Background(), "wallet-1", "USD")
	assert.Equal(t, 0.0, available)
	assert.Equal(t, 0.0, reserved)
}

func TestMemoryStorage_Reset(t *testing.T) {
	store := NewMemoryStorage()

	// Add some data
	store.SetBalance(context.Background(), "wallet-1", "USD", 100.0, 10.0)
	account := &models.SubAccount{
		ID:        "sub1",
		WalletID:  "wallet1",
		AccountID: "acc1",
	}
	store.SaveSubAccount(context.Background(), account)

	// Reset
	err := store.Reset(context.Background())
	assert.NoError(t, err)

	// Verify data is cleared
	available, reserved, _ := store.GetBalance(context.Background(), "wallet-1", "USD")
	assert.Equal(t, 0.0, available)
	assert.Equal(t, 0.0, reserved)

	_, err = store.GetSubAccount(context.Background(), "acc1")
	assert.Error(t, err)
}

func TestMemoryStorage_UpdateSubAccount(t *testing.T) {
	store := NewMemoryStorage()
	account := &models.SubAccount{
		ID:        "sub1",
		WalletID:  "wallet1",
		AccountID: "acc1",
	}

	store.SaveSubAccount(context.Background(), account)

	// Update with new values
	updated := &models.SubAccount{
		ID:        "sub1",
		WalletID:  "wallet1",
		AccountID: "acc1",
		FirstName: "John",
		LastName:  "Doe",
		Email:     "john@example.com",
	}

	err := store.UpdateSubAccount(context.Background(), updated)
	assert.NoError(t, err)

	retrieved, _ := store.GetSubAccount(context.Background(), "acc1")
	assert.Equal(t, "John", retrieved.FirstName)
	assert.Equal(t, "Doe", retrieved.LastName)
	assert.Equal(t, "john@example.com", retrieved.Email)
}
