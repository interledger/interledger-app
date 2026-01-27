package storage

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"gitlab.com/fynbos/mockxago/internal/models"
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

	balance, err := store.GetBalance(context.Background(), "wallet1", "USD")
	assert.NoError(t, err)
	assert.Equal(t, 0.0, balance)
}

func TestMemoryStorage_SetBalance(t *testing.T) {
	store := NewMemoryStorage()

	err := store.SetBalance(context.Background(), "wallet1", "USD", 100.0)
	assert.NoError(t, err)

	balance, _ := store.GetBalance(context.Background(), "wallet1", "USD")
	assert.Equal(t, 100.0, balance)
}

func TestMemoryStorage_AddBalance(t *testing.T) {
	store := NewMemoryStorage()

	store.SetBalance(context.Background(), "wallet1", "USD", 100.0)
	err := store.AddBalance(context.Background(), "wallet1", "USD", 50.0)
	assert.NoError(t, err)

	balance, _ := store.GetBalance(context.Background(), "wallet1", "USD")
	assert.Equal(t, 150.0, balance)
}

func TestMemoryStorage_SubtractBalance(t *testing.T) {
	store := NewMemoryStorage()

	store.SetBalance(context.Background(), "wallet1", "USD", 100.0)
	err := store.SubtractBalance(context.Background(), "wallet1", "USD", 30.0)
	assert.NoError(t, err)

	balance, _ := store.GetBalance(context.Background(), "wallet1", "USD")
	assert.Equal(t, 70.0, balance)
}

func TestMemoryStorage_SubtractBalance_InsufficientFunds(t *testing.T) {
	store := NewMemoryStorage()

	store.SetBalance(context.Background(), "wallet1", "USD", 50.0)
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

func TestMemoryStorage_ListBeneficiariesByWallet(t *testing.T) {
	store := NewMemoryStorage()

	for i := 0; i < 5; i++ {
		beneficiary := &models.Beneficiary{
			ID:       "ben" + string(rune(i)),
			WalletID: "wallet1",
			BankName: "Bank " + string(rune(i)),
		}
		store.SaveBeneficiary(context.Background(), beneficiary)
	}

	beneficiaries, total, err := store.ListBeneficiariesByWallet(context.Background(), "wallet1", 3, 0)
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
