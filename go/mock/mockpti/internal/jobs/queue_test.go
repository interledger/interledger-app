package jobs

import (
	"context"
	"testing"
	"time"

	"github.com/interledger/interledger-app/go/mock/mockpti/internal/storage"
)

func TestQueue_Enqueue(t *testing.T) {
	store := storage.NewMemoryStorage()
	q := NewQueue(store)

	jobID, err := q.Enqueue(JobTypeUserAssessmentWebhook, map[string]interface{}{
		"user_id":    "user-1",
		"request_id": "req-1",
	}, time.Now())
	if err != nil {
		t.Fatalf("Enqueue failed: %v", err)
	}
	if jobID == "" {
		t.Fatal("expected non-empty job ID")
	}

	job, err := store.GetJob(context.Background(), jobID)
	if err != nil {
		t.Fatalf("GetJob failed: %v", err)
	}
	if job.Status != JobStatusQueued {
		t.Errorf("expected status %s, got %s", JobStatusQueued, job.Status)
	}
}

func TestQueue_GetReadyJobs_ReturnsOnlyQueuedReady(t *testing.T) {
	store := storage.NewMemoryStorage()
	q := NewQueue(store)

	pastJobID, err := q.Enqueue("ready.job", map[string]interface{}{}, time.Now().Add(-1*time.Minute))
	if err != nil {
		t.Fatalf("enqueue past job failed: %v", err)
	}
	_, err = q.Enqueue("future.job", map[string]interface{}{}, time.Now().Add(1*time.Hour))
	if err != nil {
		t.Fatalf("enqueue future job failed: %v", err)
	}

	if err := q.MarkProcessing(pastJobID); err != nil {
		t.Fatalf("MarkProcessing failed: %v", err)
	}

	jobs, err := q.GetReadyJobs(10)
	if err != nil {
		t.Fatalf("GetReadyJobs failed: %v", err)
	}
	if len(jobs) != 0 {
		t.Fatalf("expected 0 queued-ready jobs, got %d", len(jobs))
	}
}

func TestQueue_MarkDelivered(t *testing.T) {
	store := storage.NewMemoryStorage()
	q := NewQueue(store)

	jobID, err := q.Enqueue("job", map[string]interface{}{}, time.Now())
	if err != nil {
		t.Fatalf("Enqueue failed: %v", err)
	}
	if err := q.MarkDelivered(jobID); err != nil {
		t.Fatalf("MarkDelivered failed: %v", err)
	}

	job, err := store.GetJob(context.Background(), jobID)
	if err != nil {
		t.Fatalf("GetJob failed: %v", err)
	}
	if job.Status != JobStatusDelivered {
		t.Errorf("expected status %s, got %s", JobStatusDelivered, job.Status)
	}
}

func TestQueue_MarkProcessing(t *testing.T) {
	store := storage.NewMemoryStorage()
	q := NewQueue(store)

	jobID, err := q.Enqueue("job", map[string]interface{}{}, time.Now())
	if err != nil {
		t.Fatalf("Enqueue failed: %v", err)
	}
	if err := q.MarkProcessing(jobID); err != nil {
		t.Fatalf("MarkProcessing failed: %v", err)
	}

	job, err := store.GetJob(context.Background(), jobID)
	if err != nil {
		t.Fatalf("GetJob failed: %v", err)
	}
	if job.Status != JobStatusProcessing {
		t.Errorf("expected status %s, got %s", JobStatusProcessing, job.Status)
	}
}

func TestQueue_MarkFailed_Reschedules(t *testing.T) {
	store := storage.NewMemoryStorage()
	q := NewQueue(store)

	jobID, err := q.Enqueue("job", map[string]interface{}{}, time.Now())
	if err != nil {
		t.Fatalf("Enqueue failed: %v", err)
	}

	if err := q.MarkProcessing(jobID); err != nil {
		t.Fatalf("MarkProcessing failed: %v", err)
	}
	if err := q.MarkFailed(jobID, "temporary failure"); err != nil {
		t.Fatalf("MarkFailed failed: %v", err)
	}

	job, err := store.GetJob(context.Background(), jobID)
	if err != nil {
		t.Fatalf("GetJob failed: %v", err)
	}
	if job.Status != JobStatusQueued {
		t.Errorf("expected status %s, got %s", JobStatusQueued, job.Status)
	}
	if job.Attempts != 1 {
		t.Errorf("expected attempts 1, got %d", job.Attempts)
	}
}

func TestQueue_MarkFailed_Terminal(t *testing.T) {
	store := storage.NewMemoryStorage()
	q := NewQueue(store)

	jobID, err := q.Enqueue("job", map[string]interface{}{}, time.Now())
	if err != nil {
		t.Fatalf("Enqueue failed: %v", err)
	}

	for i := 0; i < maxAttempts; i++ {
		if err := q.MarkFailed(jobID, "still failing"); err != nil {
			t.Fatalf("MarkFailed failed on attempt %d: %v", i+1, err)
		}
	}

	job, err := store.GetJob(context.Background(), jobID)
	if err != nil {
		t.Fatalf("GetJob failed: %v", err)
	}
	if job.Status != JobStatusFailed {
		t.Errorf("expected status %s, got %s", JobStatusFailed, job.Status)
	}
}
