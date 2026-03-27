package storage

import (
	"context"
	"testing"
	"time"

	miniredis "github.com/alicebob/miniredis/v2"

	"gitlab.com/fynbos/mock/mockpti/internal/models"
)

func TestRedisStorage_CoreOperations(t *testing.T) {
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("failed to start miniredis: %v", err)
	}
	defer mr.Close()

	store, err := NewRedisStorage("redis://"+mr.Addr(), 0)
	if err != nil {
		t.Fatalf("failed to create redis storage: %v", err)
	}
	ctx := context.Background()

	user := &models.User{ID: "user-1", Type: "PERSON"}
	if err := store.SaveUser(ctx, user); err != nil {
		t.Fatalf("SaveUser failed: %v", err)
	}
	if _, err := store.GetUser(ctx, user.ID); err != nil {
		t.Fatalf("GetUser failed: %v", err)
	}

	user.Type = "INDIVIDUAL"
	if err := store.UpdateUser(ctx, user); err != nil {
		t.Fatalf("UpdateUser failed: %v", err)
	}

	assessment := &models.Assessment{RequestID: "req-1", UserID: user.ID, Assessment: "ACCEPTED", Tier: 1}
	if err := store.SaveAssessment(ctx, assessment); err != nil {
		t.Fatalf("SaveAssessment failed: %v", err)
	}
	if _, err := store.GetLatestAssessment(ctx, user.ID); err != nil {
		t.Fatalf("GetLatestAssessment failed: %v", err)
	}

	wallet := &models.Wallet{WalletID: "w-1", UserID: user.ID, Currency: "USD"}
	if err := store.SaveWallet(ctx, wallet); err != nil {
		t.Fatalf("SaveWallet failed: %v", err)
	}
	if _, err := store.GetWallet(ctx, user.ID, wallet.WalletID); err != nil {
		t.Fatalf("GetWallet failed: %v", err)
	}
	wallets, err := store.ListWallets(ctx, user.ID)
	if err != nil {
		t.Fatalf("ListWallets failed: %v", err)
	}
	if len(wallets) != 1 {
		t.Fatalf("expected 1 wallet, got %d", len(wallets))
	}

	pi := &models.PaymentInformation{ID: "pi-1", UserID: user.ID, Type: "BANK_ACCOUNT"}
	if err := store.SavePaymentInformation(ctx, pi); err != nil {
		t.Fatalf("SavePaymentInformation failed: %v", err)
	}
	if _, err := store.GetPaymentInformation(ctx, user.ID, pi.ID); err != nil {
		t.Fatalf("GetPaymentInformation failed: %v", err)
	}

	tx := &models.Transaction{RequestID: "tx-1", Status: "SETTLED", TransactionType: "DEPOSIT"}
	if err := store.SaveTransaction(ctx, tx); err != nil {
		t.Fatalf("SaveTransaction failed: %v", err)
	}
	if _, err := store.GetTransaction(ctx, tx.RequestID); err != nil {
		t.Fatalf("GetTransaction failed: %v", err)
	}
	update := &models.TransactionUpdate{ID: "u-1", RequestID: tx.RequestID, TransactionID: tx.RequestID}
	if err := store.SaveTransactionUpdate(ctx, update); err != nil {
		t.Fatalf("SaveTransactionUpdate failed: %v", err)
	}

	job := &models.Job{ID: "job-1", JobType: "webhook", Status: "queued", NotBefore: time.Now().Add(-1 * time.Second)}
	if err := store.SaveJob(ctx, job); err != nil {
		t.Fatalf("SaveJob failed: %v", err)
	}
	if _, err := store.GetJob(ctx, job.ID); err != nil {
		t.Fatalf("GetJob failed: %v", err)
	}
	ready, err := store.ListReadyJobs(ctx, 10)
	if err != nil {
		t.Fatalf("ListReadyJobs failed: %v", err)
	}
	if len(ready) != 1 {
		t.Fatalf("expected 1 ready job, got %d", len(ready))
	}
	if err := store.IncrementJobAttempts(ctx, job.ID); err != nil {
		t.Fatalf("IncrementJobAttempts failed: %v", err)
	}
	now := time.Now()
	if err := store.UpdateJobStatus(ctx, job.ID, "delivered", &now, ""); err != nil {
		t.Fatalf("UpdateJobStatus failed: %v", err)
	}
	if err := store.ClearJobs(ctx); err != nil {
		t.Fatalf("ClearJobs failed: %v", err)
	}

	if err := store.Reset(ctx); err != nil {
		t.Fatalf("Reset failed: %v", err)
	}
}

func TestRedisStorage_NotFoundErrors(t *testing.T) {
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("failed to start miniredis: %v", err)
	}
	defer mr.Close()

	store, err := NewRedisStorage("redis://"+mr.Addr(), 0)
	if err != nil {
		t.Fatalf("failed to create redis storage: %v", err)
	}
	ctx := context.Background()

	if _, err := store.GetUser(ctx, "missing"); err != ErrUserNotFound {
		t.Fatalf("expected ErrUserNotFound, got %v", err)
	}
	if _, err := store.GetLatestAssessment(ctx, "missing"); err != ErrAssessmentNotFound {
		t.Fatalf("expected ErrAssessmentNotFound, got %v", err)
	}
	if _, err := store.GetWallet(ctx, "missing", "missing"); err != ErrWalletNotFound {
		t.Fatalf("expected ErrWalletNotFound, got %v", err)
	}
	if _, err := store.GetPaymentInformation(ctx, "missing", "missing"); err != ErrPaymentInformationNotFound {
		t.Fatalf("expected ErrPaymentInformationNotFound, got %v", err)
	}
	if _, err := store.GetTransaction(ctx, "missing"); err != ErrTransactionNotFound {
		t.Fatalf("expected ErrTransactionNotFound, got %v", err)
	}
	if _, err := store.GetJob(ctx, "missing"); err != ErrJobNotFound {
		t.Fatalf("expected ErrJobNotFound, got %v", err)
	}
}
