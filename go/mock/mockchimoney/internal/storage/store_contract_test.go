package storage

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"

	"gitlab.com/fynbos/mock/mockchimoney/internal/models"
)

type storeFactory func(t *testing.T) Store

func runStoreContractTests(t *testing.T, newStore storeFactory) {
	t.Helper()

	t.Run("CreateAndGetSubAccount", func(t *testing.T) {
		t.Parallel()

		store := newStore(t)
		account := models.SubAccount{ID: uuid.NewString(), Name: "Alice"}

		created, err := store.CreateSubAccount(context.Background(), account)
		if err != nil {
			t.Fatalf("CreateSubAccount() unexpected error: %v", err)
		}
		if created.ID != account.ID {
			t.Fatalf("CreateSubAccount() ID mismatch: got %q want %q", created.ID, account.ID)
		}

		got, err := store.GetSubAccount(context.Background(), account.ID)
		if err != nil {
			t.Fatalf("GetSubAccount() unexpected error: %v", err)
		}
		if got.ID != account.ID {
			t.Fatalf("GetSubAccount() ID mismatch: got %q want %q", got.ID, account.ID)
		}
	})

	t.Run("PaymentLifecycle", func(t *testing.T) {
		t.Parallel()

		store := newStore(t)
		p := models.Payment{ID: uuid.NewString(), IssueID: "a_b", Amount: 10, Currency: "CAD", Status: "pending", CreatedAt: time.Now().UTC()}

		if _, err := store.CreatePayment(context.Background(), p); err != nil {
			t.Fatalf("CreatePayment() unexpected error: %v", err)
		}

		got, err := store.GetPaymentByIssueID(context.Background(), p.IssueID)
		if err != nil {
			t.Fatalf("GetPaymentByIssueID() unexpected error: %v", err)
		}
		if got.Status != "pending" {
			t.Fatalf("status mismatch: got %q want %q", got.Status, "pending")
		}

		updated, err := store.UpdatePaymentStatus(context.Background(), p.IssueID, "redeemed")
		if err != nil {
			t.Fatalf("UpdatePaymentStatus() unexpected error: %v", err)
		}
		if updated.Status != "redeemed" {
			t.Fatalf("updated status mismatch: got %q", updated.Status)
		}
	})

	t.Run("PayoutLifecycle", func(t *testing.T) {
		t.Parallel()

		store := newStore(t)
		p := models.Payout{ID: uuid.NewString(), IssueID: "x_y", ChiRef: uuid.NewString(), Amount: 22, Currency: "CAD", Status: "pending", CreatedAt: time.Now().UTC()}

		if _, err := store.CreatePayout(context.Background(), p); err != nil {
			t.Fatalf("CreatePayout() unexpected error: %v", err)
		}

		got, err := store.GetPayoutByChiRef(context.Background(), p.ChiRef)
		if err != nil {
			t.Fatalf("GetPayoutByChiRef() unexpected error: %v", err)
		}
		if got.IssueID != p.IssueID {
			t.Fatalf("issue mismatch: got %q want %q", got.IssueID, p.IssueID)
		}

		if _, err := store.GetPayoutByIssueID(context.Background(), p.IssueID); err != nil {
			t.Fatalf("GetPayoutByIssueID() unexpected error: %v", err)
		}

		updated, err := store.UpdatePayoutStatus(context.Background(), p.IssueID, "completed")
		if err != nil {
			t.Fatalf("UpdatePayoutStatus() unexpected error: %v", err)
		}
		if updated.Status != "completed" {
			t.Fatalf("status mismatch: got %q", updated.Status)
		}
	})

	t.Run("UpdateSubAccountKYCStatus", func(t *testing.T) {
		t.Parallel()

		store := newStore(t)
		acct := models.SubAccount{ID: "kyc-1", Name: "A", KYCStatus: "pending"}
		if _, err := store.CreateSubAccount(context.Background(), acct); err != nil {
			t.Fatalf("CreateSubAccount() unexpected error: %v", err)
		}

		updated, err := store.UpdateSubAccountKYCStatus(context.Background(), acct.ID, "completed")
		if err != nil {
			t.Fatalf("UpdateSubAccountKYCStatus() unexpected error: %v", err)
		}
		if updated.KYCStatus != "completed" {
			t.Fatalf("KYC status mismatch: got %q", updated.KYCStatus)
		}
	})

	t.Run("GetSubAccountNotFound", func(t *testing.T) {
		t.Parallel()

		store := newStore(t)
		_, err := store.GetSubAccount(context.Background(), "missing")
		if err != ErrNotFound {
			t.Fatalf("GetSubAccount() error mismatch: got %v want %v", err, ErrNotFound)
		}
	})

	t.Run("ListSubAccounts", func(t *testing.T) {
		t.Parallel()

		store := newStore(t)
		ids := []string{uuid.NewString(), uuid.NewString()}
		for _, id := range ids {
			_, err := store.CreateSubAccount(context.Background(), models.SubAccount{ID: id, Name: id})
			if err != nil {
				t.Fatalf("CreateSubAccount() unexpected error: %v", err)
			}
		}

		accounts, err := store.ListSubAccounts(context.Background())
		if err != nil {
			t.Fatalf("ListSubAccounts() unexpected error: %v", err)
		}
		if len(accounts) != 2 {
			t.Fatalf("ListSubAccounts() count mismatch: got %d want 2", len(accounts))
		}
	})
}
