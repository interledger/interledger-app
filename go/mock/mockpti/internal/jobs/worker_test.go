package jobs

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"gitlab.com/fynbos/mock/mockpti/internal/models"
	"gitlab.com/fynbos/mock/mockpti/internal/storage"
)

func TestWorker_ProcessesReadyJob(t *testing.T) {
	store := storage.NewMemoryStorage()
	q := NewQueue(store)
	w := NewWorker(q)

	var called int32
	w.RegisterHandler("test.job", func(ctx context.Context, job *models.Job) error {
		atomic.StoreInt32(&called, 1)
		return nil
	})

	jobID, err := q.Enqueue("test.job", map[string]interface{}{}, time.Now())
	if err != nil {
		t.Fatalf("Enqueue failed: %v", err)
	}

	// Process one batch directly
	w.processReadyJobs()

	if atomic.LoadInt32(&called) != 1 {
		t.Error("expected handler to be called")
	}

	ctx := context.Background()
	job, err := store.GetJob(ctx, jobID)
	if err != nil {
		t.Fatalf("GetJob failed: %v", err)
	}
	if job.Status != JobStatusDelivered {
		t.Errorf("expected status %s, got %s", JobStatusDelivered, job.Status)
	}
}

func TestWorker_HandlerError_RetiesJob(t *testing.T) {
	store := storage.NewMemoryStorage()
	q := NewQueue(store)
	w := NewWorker(q)

	w.RegisterHandler("failing.job", func(ctx context.Context, job *models.Job) error {
		return errors.New("transient failure")
	})

	jobID, err := q.Enqueue("failing.job", map[string]interface{}{}, time.Now())
	if err != nil {
		t.Fatalf("Enqueue failed: %v", err)
	}

	w.processReadyJobs()

	ctx := context.Background()
	job, err := store.GetJob(ctx, jobID)
	if err != nil {
		t.Fatalf("GetJob failed: %v", err)
	}
	// Should be rescheduled (still queued)
	if job.Status != JobStatusQueued {
		t.Errorf("expected status %s (rescheduled), got %s", JobStatusQueued, job.Status)
	}
	if job.Attempts != 1 {
		t.Errorf("expected 1 attempt, got %d", job.Attempts)
	}
}

func TestWorker_UnknownJobType_PermanentlyFails(t *testing.T) {
	store := storage.NewMemoryStorage()
	q := NewQueue(store)
	w := NewWorker(q)
	// No handler registered

	jobID, err := q.Enqueue("unknown.type", map[string]interface{}{}, time.Now())
	if err != nil {
		t.Fatalf("Enqueue failed: %v", err)
	}

	// Process until permanently failed (no handler means immediate permanent fail attempt exhaustion)
	// Each call: increments attempts and reschedules; after maxAttempts => failed
	for i := 0; i < maxAttempts+1; i++ {
		// Force the job to be immediately ready
		ctx := context.Background()
		job, _ := store.GetJob(ctx, jobID)
		if job != nil && job.Status == JobStatusQueued {
			job.NotBefore = time.Now().Add(-1 * time.Second)
			_ = store.SaveJob(ctx, job)
		}
		w.processReadyJobs()
	}

	ctx := context.Background()
	job, err := store.GetJob(ctx, jobID)
	if err != nil {
		t.Fatalf("GetJob failed: %v", err)
	}
	if job.Status != JobStatusFailed {
		t.Errorf("expected status %s, got %s", JobStatusFailed, job.Status)
	}
}

func TestWorker_FutureJobNotProcessed(t *testing.T) {
	store := storage.NewMemoryStorage()
	q := NewQueue(store)
	w := NewWorker(q)

	var called int32
	w.RegisterHandler("future.job", func(ctx context.Context, job *models.Job) error {
		atomic.StoreInt32(&called, 1)
		return nil
	})

	_, err := q.Enqueue("future.job", map[string]interface{}{}, time.Now().Add(1*time.Hour))
	if err != nil {
		t.Fatalf("Enqueue failed: %v", err)
	}

	w.processReadyJobs()

	if atomic.LoadInt32(&called) != 0 {
		t.Error("future job should not have been processed yet")
	}
}

func TestWorker_DataPassedToHandler(t *testing.T) {
	store := storage.NewMemoryStorage()
	q := NewQueue(store)
	w := NewWorker(q)

	var gotUserID string
	w.RegisterHandler("data.job", func(ctx context.Context, job *models.Job) error {
		gotUserID, _ = job.Data["user_id"].(string)
		return nil
	})

	_, err := q.Enqueue("data.job", map[string]interface{}{"user_id": "user-42"}, time.Now())
	if err != nil {
		t.Fatalf("Enqueue failed: %v", err)
	}

	w.processReadyJobs()

	if gotUserID != "user-42" {
		t.Errorf("expected user_id 'user-42', got %q", gotUserID)
	}
}
