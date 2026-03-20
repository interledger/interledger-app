package storage

import (
	"context"
	"testing"

	"gitlab.com/fynbos/mock/mockpti/internal/models"
)

func TestMemoryStorage_SaveAndGetUser(t *testing.T) {
	store := NewMemoryStorage()
	ctx := context.Background()

	user := &models.User{
		ID:   "user-1",
		Type: "PERSON",
		Name: &models.Name{First: "Alice", Last: "Smith"},
	}

	if err := store.SaveUser(ctx, user); err != nil {
		t.Fatalf("SaveUser failed: %v", err)
	}

	got, err := store.GetUser(ctx, "user-1")
	if err != nil {
		t.Fatalf("GetUser failed: %v", err)
	}

	if got.ID != "user-1" {
		t.Errorf("expected ID user-1, got %s", got.ID)
	}
	if got.Name.First != "Alice" {
		t.Errorf("expected first name Alice, got %s", got.Name.First)
	}
}

func TestMemoryStorage_GetUser_NotFound(t *testing.T) {
	store := NewMemoryStorage()
	ctx := context.Background()

	_, err := store.GetUser(ctx, "nonexistent")
	if err != ErrUserNotFound {
		t.Errorf("expected ErrUserNotFound, got %v", err)
	}
}

func TestMemoryStorage_UpdateUser(t *testing.T) {
	store := NewMemoryStorage()
	ctx := context.Background()

	user := &models.User{
		ID:   "user-1",
		Type: "PERSON",
		Name: &models.Name{First: "Alice", Last: "Smith"},
	}
	if err := store.SaveUser(ctx, user); err != nil {
		t.Fatalf("SaveUser failed: %v", err)
	}

	user.Name.First = "Bob"
	if err := store.UpdateUser(ctx, user); err != nil {
		t.Fatalf("UpdateUser failed: %v", err)
	}

	got, err := store.GetUser(ctx, "user-1")
	if err != nil {
		t.Fatalf("GetUser failed: %v", err)
	}
	if got.Name.First != "Bob" {
		t.Errorf("expected first name Bob, got %s", got.Name.First)
	}
}

func TestMemoryStorage_UpdateUser_NotFound(t *testing.T) {
	store := NewMemoryStorage()
	ctx := context.Background()

	user := &models.User{ID: "nonexistent"}
	err := store.UpdateUser(ctx, user)
	if err != ErrUserNotFound {
		t.Errorf("expected ErrUserNotFound, got %v", err)
	}
}

func TestMemoryStorage_SaveAndGetAssessment(t *testing.T) {
	store := NewMemoryStorage()
	ctx := context.Background()

	assessment := &models.Assessment{
		ResourceType: "assessment",
		RequestID:    "req-1",
		UserID:       "user-1",
		Assessment:   "approved",
		Tier:         1,
	}

	if err := store.SaveAssessment(ctx, assessment); err != nil {
		t.Fatalf("SaveAssessment failed: %v", err)
	}

	got, err := store.GetLatestAssessment(ctx, "user-1")
	if err != nil {
		t.Fatalf("GetLatestAssessment failed: %v", err)
	}

	if got.RequestID != "req-1" {
		t.Errorf("expected request ID req-1, got %s", got.RequestID)
	}
	if got.Assessment != "approved" {
		t.Errorf("expected assessment approved, got %s", got.Assessment)
	}
}

func TestMemoryStorage_GetLatestAssessment_ReturnsNewest(t *testing.T) {
	store := NewMemoryStorage()
	ctx := context.Background()

	a1 := &models.Assessment{RequestID: "req-1", UserID: "user-1", Assessment: "pending"}
	a2 := &models.Assessment{RequestID: "req-2", UserID: "user-1", Assessment: "approved"}

	_ = store.SaveAssessment(ctx, a1)
	_ = store.SaveAssessment(ctx, a2)

	got, err := store.GetLatestAssessment(ctx, "user-1")
	if err != nil {
		t.Fatalf("GetLatestAssessment failed: %v", err)
	}
	if got.RequestID != "req-2" {
		t.Errorf("expected latest assessment req-2, got %s", got.RequestID)
	}
}

func TestMemoryStorage_GetLatestAssessment_NotFound(t *testing.T) {
	store := NewMemoryStorage()
	ctx := context.Background()

	_, err := store.GetLatestAssessment(ctx, "nonexistent")
	if err != ErrAssessmentNotFound {
		t.Errorf("expected ErrAssessmentNotFound, got %v", err)
	}
}

func TestMemoryStorage_Reset(t *testing.T) {
	store := NewMemoryStorage()
	ctx := context.Background()

	_ = store.SaveUser(ctx, &models.User{ID: "user-1"})
	_ = store.SaveAssessment(ctx, &models.Assessment{UserID: "user-1", RequestID: "req-1"})
	_ = store.SaveWallet(ctx, &models.Wallet{WalletID: "w1", UserID: "user-1", Currency: "USD"})
	_ = store.SavePaymentInformation(ctx, &models.PaymentInformation{ID: "pi1", UserID: "user-1", Type: "BANK_ACCOUNT"})

	if err := store.Reset(ctx); err != nil {
		t.Fatalf("Reset failed: %v", err)
	}

	_, err := store.GetUser(ctx, "user-1")
	if err != ErrUserNotFound {
		t.Errorf("expected ErrUserNotFound after reset, got %v", err)
	}

	_, err = store.GetLatestAssessment(ctx, "user-1")
	if err != ErrAssessmentNotFound {
		t.Errorf("expected ErrAssessmentNotFound after reset, got %v", err)
	}

	_, err = store.GetWallet(ctx, "user-1", "w1")
	if err != ErrWalletNotFound {
		t.Errorf("expected ErrWalletNotFound after reset, got %v", err)
	}

	_, err = store.GetPaymentInformation(ctx, "user-1", "pi1")
	if err != ErrPaymentInformationNotFound {
		t.Errorf("expected ErrPaymentInformationNotFound after reset, got %v", err)
	}
}

// Wallet storage tests

func TestMemoryStorage_SaveAndGetWallet(t *testing.T) {
	store := NewMemoryStorage()
	ctx := context.Background()

	wallet := &models.Wallet{
		WalletID:       "wallet-1",
		Currency:       "USD",
		Reference:      "test-ref",
		CreateDateTime: "2026-03-20T00:00:00Z",
		Balance:        0,
		UserID:         "user-1",
	}

	if err := store.SaveWallet(ctx, wallet); err != nil {
		t.Fatalf("SaveWallet failed: %v", err)
	}

	got, err := store.GetWallet(ctx, "user-1", "wallet-1")
	if err != nil {
		t.Fatalf("GetWallet failed: %v", err)
	}

	if got.WalletID != "wallet-1" {
		t.Errorf("expected walletId wallet-1, got %s", got.WalletID)
	}
	if got.Currency != "USD" {
		t.Errorf("expected currency USD, got %s", got.Currency)
	}
	if got.Reference != "test-ref" {
		t.Errorf("expected reference test-ref, got %s", got.Reference)
	}
}

func TestMemoryStorage_GetWallet_NotFound(t *testing.T) {
	store := NewMemoryStorage()
	ctx := context.Background()

	_, err := store.GetWallet(ctx, "user-1", "nonexistent")
	if err != ErrWalletNotFound {
		t.Errorf("expected ErrWalletNotFound, got %v", err)
	}
}

func TestMemoryStorage_GetWallet_WrongUser(t *testing.T) {
	store := NewMemoryStorage()
	ctx := context.Background()

	_ = store.SaveWallet(ctx, &models.Wallet{WalletID: "w1", UserID: "user-1", Currency: "USD"})

	_, err := store.GetWallet(ctx, "user-2", "w1")
	if err != ErrWalletNotFound {
		t.Errorf("expected ErrWalletNotFound for wrong user, got %v", err)
	}
}

func TestMemoryStorage_ListWallets(t *testing.T) {
	store := NewMemoryStorage()
	ctx := context.Background()

	_ = store.SaveWallet(ctx, &models.Wallet{WalletID: "w1", UserID: "user-1", Currency: "USD"})
	_ = store.SaveWallet(ctx, &models.Wallet{WalletID: "w2", UserID: "user-1", Currency: "EUR"})
	_ = store.SaveWallet(ctx, &models.Wallet{WalletID: "w3", UserID: "user-2", Currency: "GBP"})

	wallets, err := store.ListWallets(ctx, "user-1")
	if err != nil {
		t.Fatalf("ListWallets failed: %v", err)
	}

	if len(wallets) != 2 {
		t.Errorf("expected 2 wallets for user-1, got %d", len(wallets))
	}
}

func TestMemoryStorage_ListWallets_Empty(t *testing.T) {
	store := NewMemoryStorage()
	ctx := context.Background()

	wallets, err := store.ListWallets(ctx, "user-no-wallets")
	if err != nil {
		t.Fatalf("ListWallets failed: %v", err)
	}

	if len(wallets) != 0 {
		t.Errorf("expected 0 wallets, got %d", len(wallets))
	}
}

// Payment information storage tests

func TestMemoryStorage_SaveAndGetPaymentInformation(t *testing.T) {
	store := NewMemoryStorage()
	ctx := context.Background()

	pi := &models.PaymentInformation{
		ID:                "pi-1",
		Type:              "BANK_ACCOUNT",
		BankAccountNumber: "123456789",
		BankRoutingNumber: "021000021",
		AccountBankName:   "Test Bank",
		UserID:            "user-1",
	}

	if err := store.SavePaymentInformation(ctx, pi); err != nil {
		t.Fatalf("SavePaymentInformation failed: %v", err)
	}

	got, err := store.GetPaymentInformation(ctx, "user-1", "pi-1")
	if err != nil {
		t.Fatalf("GetPaymentInformation failed: %v", err)
	}

	if got.ID != "pi-1" {
		t.Errorf("expected ID pi-1, got %s", got.ID)
	}
	if got.Type != "BANK_ACCOUNT" {
		t.Errorf("expected type BANK_ACCOUNT, got %s", got.Type)
	}
	if got.BankAccountNumber != "123456789" {
		t.Errorf("expected bank account 123456789, got %s", got.BankAccountNumber)
	}
}

func TestMemoryStorage_GetPaymentInformation_NotFound(t *testing.T) {
	store := NewMemoryStorage()
	ctx := context.Background()

	_, err := store.GetPaymentInformation(ctx, "user-1", "nonexistent")
	if err != ErrPaymentInformationNotFound {
		t.Errorf("expected ErrPaymentInformationNotFound, got %v", err)
	}
}

func TestMemoryStorage_GetPaymentInformation_WrongUser(t *testing.T) {
	store := NewMemoryStorage()
	ctx := context.Background()

	_ = store.SavePaymentInformation(ctx, &models.PaymentInformation{ID: "pi-1", UserID: "user-1", Type: "BANK_ACCOUNT"})

	_, err := store.GetPaymentInformation(ctx, "user-2", "pi-1")
	if err != ErrPaymentInformationNotFound {
		t.Errorf("expected ErrPaymentInformationNotFound for wrong user, got %v", err)
	}
}
