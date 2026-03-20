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
}
