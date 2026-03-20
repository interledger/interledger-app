package storage

import (
	"context"
	"testing"

	"github.com/google/uuid"

	"gitlab.com/fynbos/mock/mockchimoney/internal/models"
)

func TestMemoryStore_CreateAndGetSubAccount(t *testing.T) {
	t.Parallel()

	store := NewMemoryStore()
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
}

func TestMemoryStore_GetSubAccountNotFound(t *testing.T) {
	t.Parallel()

	store := NewMemoryStore()
	_, err := store.GetSubAccount(context.Background(), "missing")
	if err != ErrNotFound {
		t.Fatalf("GetSubAccount() error mismatch: got %v want %v", err, ErrNotFound)
	}
}

func TestMemoryStore_ListSubAccounts(t *testing.T) {
	t.Parallel()

	store := NewMemoryStore()
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
}
