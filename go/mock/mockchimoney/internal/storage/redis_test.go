package storage

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/google/uuid"

	"gitlab.com/fynbos/mock/mockchimoney/internal/models"
)

func TestRedisStore_Contract(t *testing.T) {
	runStoreContractTests(t, func(t *testing.T) Store {
		t.Helper()

		mini, err := miniredis.Run()
		if err != nil {
			t.Fatalf("miniredis.Run() unexpected error: %v", err)
		}
		t.Cleanup(mini.Close)

		store, err := NewRedisStore("redis://"+mini.Addr(), 0)
		if err != nil {
			t.Fatalf("NewRedisStore() unexpected error: %v", err)
		}
		t.Cleanup(func() {
			if closeErr := store.Close(); closeErr != nil {
				t.Fatalf("Close() unexpected error: %v", closeErr)
			}
		})
		return store
	})
}

func TestRedisStore_PersistsAcrossInstances(t *testing.T) {
	mini, err := miniredis.Run()
	if err != nil {
		t.Fatalf("miniredis.Run() unexpected error: %v", err)
	}
	defer mini.Close()

	redisURL := "redis://" + mini.Addr()

	store1, err := NewRedisStore(redisURL, 0)
	if err != nil {
		t.Fatalf("NewRedisStore() unexpected error: %v", err)
	}

	sub := models.SubAccount{ID: uuid.NewString(), Name: "persist-sub", CreatedAt: time.Now().UTC()}
	pay := models.Payment{ID: uuid.NewString(), IssueID: "persist_issue", Status: "pending", CreatedAt: time.Now().UTC()}
	payout := models.Payout{ID: uuid.NewString(), IssueID: "persist_payout_issue", ChiRef: uuid.NewString(), Status: "pending", CreatedAt: time.Now().UTC()}

	if _, err := store1.CreateSubAccount(context.Background(), sub); err != nil {
		t.Fatalf("CreateSubAccount() unexpected error: %v", err)
	}
	if _, err := store1.CreatePayment(context.Background(), pay); err != nil {
		t.Fatalf("CreatePayment() unexpected error: %v", err)
	}
	if _, err := store1.CreatePayout(context.Background(), payout); err != nil {
		t.Fatalf("CreatePayout() unexpected error: %v", err)
	}

	if err := store1.Close(); err != nil {
		t.Fatalf("Close() unexpected error: %v", err)
	}

	store2, err := NewRedisStore(redisURL, 0)
	if err != nil {
		t.Fatalf("NewRedisStore() second instance unexpected error: %v", err)
	}
	defer func() {
		if closeErr := store2.Close(); closeErr != nil {
			t.Fatalf("Close() second instance unexpected error: %v", closeErr)
		}
	}()

	if got, err := store2.GetSubAccount(context.Background(), sub.ID); err != nil || got.ID != sub.ID {
		t.Fatalf("GetSubAccount() after restart mismatch: got=%+v err=%v", got, err)
	}
	if got, err := store2.GetPaymentByIssueID(context.Background(), pay.IssueID); err != nil || got.IssueID != pay.IssueID {
		t.Fatalf("GetPaymentByIssueID() after restart mismatch: got=%+v err=%v", got, err)
	}
	if got, err := store2.GetPayoutByIssueID(context.Background(), payout.IssueID); err != nil || got.IssueID != payout.IssueID {
		t.Fatalf("GetPayoutByIssueID() after restart mismatch: got=%+v err=%v", got, err)
	}
}
