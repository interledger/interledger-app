package jobs

import (
	"context"
	"fmt"
	"time"

	"gitlab.com/fynbos/mock/mockxago/internal/logger"
	"gitlab.com/fynbos/mock/mockxago/internal/models"
	"gitlab.com/fynbos/mock/mockxago/internal/storage"

	"github.com/google/uuid"
)

const (
	maxAttempts  = 10
	retryDelay   = 30 * time.Second
	pollInterval = 1 * time.Second
	batchSize    = 10
)

// Queue is a job queue that uses storage for persistence.
type Queue struct {
	store storage.Storage
}

// NewQueue creates a new job queue with storage persistence.
func NewQueue(store storage.Storage) *Queue {
	return &Queue{store: store}
}

// Enqueue adds a new job to the queue with the given type and data.
// The job will be ready for processing at readyAt (use time.Now() for immediate).
func (q *Queue) Enqueue(jobType string, data map[string]interface{}, readyAt time.Time) (string, error) {
	ctx := context.Background()
	job := &models.Job{
		ID:        uuid.NewString(),
		JobType:   jobType,
		Data:      data,
		Attempts:  0,
		Status:    "pending",
		CreatedAt: time.Now(),
		NotBefore: readyAt,
	}

	if err := q.store.SaveJob(ctx, job); err != nil {
		return "", fmt.Errorf("failed to save job: %w", err)
	}

	logger.Infof("Job enqueued: id=%s type=%s readyAt=%s", job.ID, job.JobType, job.NotBefore.Format(time.RFC3339Nano))
	return job.ID, nil
}

// GetReadyJobs returns up to `limit` pending jobs whose NotBefore <= now,
// sorted by NotBefore ascending (oldest first).
func (q *Queue) GetReadyJobs(limit int) ([]*models.Job, error) {
	ctx := context.Background()
	jobs, err := q.store.ListReadyJobs(ctx, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to list ready jobs: %w", err)
	}
	return jobs, nil
}

// MarkCompleted marks a job as completed.
func (q *Queue) MarkCompleted(jobID string) error {
	ctx := context.Background()
	now := time.Now()
	return q.store.UpdateJobStatus(ctx, jobID, "completed", &now, "")
}

// MarkFailed records a failure. If max attempts reached, the job is permanently failed;
// otherwise it is rescheduled with a delay.
func (q *Queue) MarkFailed(jobID string, errMsg string) error {
	ctx := context.Background()

	// Increment attempts
	if err := q.store.IncrementJobAttempts(ctx, jobID); err != nil {
		return fmt.Errorf("failed to increment job attempts: %w", err)
	}

	// Get the job to check attempts
	job, err := q.store.GetJob(ctx, jobID)
	if err != nil {
		return fmt.Errorf("failed to get job: %w", err)
	}
	if job == nil {
		return fmt.Errorf("job %s not found", jobID)
	}

	if job.Attempts >= maxAttempts {
		now := time.Now()
		if err := q.store.UpdateJobStatus(ctx, jobID, "failed", &now, errMsg); err != nil {
			return fmt.Errorf("failed to mark job as failed: %w", err)
		}
		logger.Errorf("Job permanently failed: id=%s type=%s attempts=%d error=%s",
			job.ID, job.JobType, job.Attempts, errMsg)
		return nil
	}

	// Reschedule with delay - update NotBefore
	job.NotBefore = time.Now().Add(retryDelay)
	if err := q.store.SaveJob(ctx, job); err != nil {
		return fmt.Errorf("failed to reschedule job: %w", err)
	}

	logger.Warnf("Job rescheduled: id=%s type=%s attempt=%d/%d retryAt=%s error=%s",
		job.ID, job.JobType, job.Attempts, maxAttempts, job.NotBefore.Format(time.RFC3339), errMsg)
	return nil
}

// PendingCount returns the number of pending jobs in the queue.
func (q *Queue) PendingCount() int {
	ctx := context.Background()
	jobs, err := q.store.ListReadyJobs(ctx, 1000) // Get a large number to count
	if err != nil {
		logger.Errorf("Failed to count pending jobs: %v", err)
		return 0
	}
	return len(jobs)
}
