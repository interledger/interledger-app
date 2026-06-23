package storage

import (
	"context"
	"testing"
	"time"

	miniredis "github.com/alicebob/miniredis/v2"

	"github.com/interledger/interledger-app/go/mock/mockpti/internal/models"
)

func TestNewRedisStorage_AddressParsingAndConnectionError(t *testing.T) {
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("failed to start miniredis: %v", err)
	}
	defer mr.Close()

	if _, err := NewRedisStorage(mr.Addr(), 0); err != nil {
		t.Fatalf("expected storage creation with raw address to succeed, got %v", err)
	}

	if _, err := NewRedisStorage("127.0.0.1:1", 0); err == nil {
		t.Fatal("expected connection error for invalid redis address")
	}
}

func TestRedisStorage_UpdateUser_NotFound(t *testing.T) {
	mr, _ := miniredis.Run()
	defer mr.Close()

	store, _ := NewRedisStorage("redis://"+mr.Addr(), 0)
	err := store.UpdateUser(context.Background(), &models.User{ID: "missing"})
	if err != ErrUserNotFound {
		t.Fatalf("expected ErrUserNotFound, got %v", err)
	}
}

func TestRedisStorage_ListReadyJobs_FiltersByQueuedStatus(t *testing.T) {
	mr, _ := miniredis.Run()
	defer mr.Close()

	store, _ := NewRedisStorage("redis://"+mr.Addr(), 0)
	ctx := context.Background()

	job := &models.Job{ID: "job-filter", JobType: "webhook", Status: "queued", NotBefore: time.Now().Add(-1 * time.Second)}
	if err := store.SaveJob(ctx, job); err != nil {
		t.Fatalf("SaveJob failed: %v", err)
	}

	if err := store.UpdateJobStatus(ctx, job.ID, "processing", nil, ""); err != nil {
		t.Fatalf("UpdateJobStatus failed: %v", err)
	}

	ready, err := store.ListReadyJobs(ctx, 10)
	if err != nil {
		t.Fatalf("ListReadyJobs failed: %v", err)
	}
	if len(ready) != 0 {
		t.Fatalf("expected no ready jobs because status is processing, got %d", len(ready))
	}
}

func TestRedisStorage_JobMethods_NotFound(t *testing.T) {
	mr, _ := miniredis.Run()
	defer mr.Close()

	store, _ := NewRedisStorage("redis://"+mr.Addr(), 0)
	ctx := context.Background()

	if err := store.IncrementJobAttempts(ctx, "missing"); err != ErrJobNotFound {
		t.Fatalf("expected ErrJobNotFound for IncrementJobAttempts, got %v", err)
	}
	if err := store.UpdateJobStatus(ctx, "missing", "failed", nil, "oops"); err != ErrJobNotFound {
		t.Fatalf("expected ErrJobNotFound for UpdateJobStatus, got %v", err)
	}
}
