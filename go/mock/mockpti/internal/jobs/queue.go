package jobs

import (
	"context"
	"fmt"
	"time"

	"github.com/interledger/interledger-app/go/mock/mockpti/internal/logger"
	"github.com/interledger/interledger-app/go/mock/mockpti/internal/models"
	"github.com/interledger/interledger-app/go/mock/mockpti/internal/storage"
	"github.com/interledger/interledger-app/go/mock/mockpti/internal/utils"
)

const (
	maxAttempts  = 5
	retryDelay   = 10 * time.Second
	pollInterval = 500 * time.Millisecond
	batchSize    = 10
)

// Queue is a job queue backed by the Storage interface.
type Queue struct {
	store storage.Storage
}

// NewQueue creates a new queue backed by the given store.
func NewQueue(store storage.Storage) *Queue {
	return &Queue{store: store}
}

// Enqueue persists a new job with the given type and data, ready at readyAt.
// Use time.Now() for immediate processing.
func (q *Queue) Enqueue(jobType string, data map[string]interface{}, readyAt time.Time) (string, error) {
	ctx := context.Background()
	job := &models.Job{
		ID:        utils.GenerateUUID(),
		JobType:   jobType,
		Data:      data,
		Attempts:  0,
		Status:    JobStatusQueued,
		CreatedAt: time.Now(),
		NotBefore: readyAt,
	}

	if err := q.store.SaveJob(ctx, job); err != nil {
		return "", fmt.Errorf("failed to save job: %w", err)
	}

	logger.Infof("Job enqueued: id=%s type=%s readyAt=%s", job.ID, job.JobType, job.NotBefore.Format(time.RFC3339))
	return job.ID, nil
}

// GetReadyJobs returns up to limit pending jobs whose NotBefore <= now.
func (q *Queue) GetReadyJobs(limit int) ([]*models.Job, error) {
	ctx := context.Background()
	jobs, err := q.store.ListReadyJobs(ctx, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to list ready jobs: %w", err)
	}
	return jobs, nil
}

// MarkDelivered marks a job as successfully delivered.
func (q *Queue) MarkDelivered(jobID string) error {
	ctx := context.Background()
	now := time.Now()
	return q.store.UpdateJobStatus(ctx, jobID, JobStatusDelivered, &now, "")
}

// MarkProcessing marks a job as actively being processed.
func (q *Queue) MarkProcessing(jobID string) error {
	ctx := context.Background()
	return q.store.UpdateJobStatus(ctx, jobID, JobStatusProcessing, nil, "")
}

// MarkFailed records a failure attempt. If max attempts reached, the job is
// permanently failed; otherwise it is rescheduled with exponential backoff.
func (q *Queue) MarkFailed(jobID string, errMsg string) error {
	ctx := context.Background()

	if err := q.store.IncrementJobAttempts(ctx, jobID); err != nil {
		return fmt.Errorf("failed to increment job attempts: %w", err)
	}

	job, err := q.store.GetJob(ctx, jobID)
	if err != nil {
		return fmt.Errorf("failed to get job: %w", err)
	}

	if job.Attempts >= maxAttempts {
		now := time.Now()
		if err := q.store.UpdateJobStatus(ctx, jobID, JobStatusFailed, &now, errMsg); err != nil {
			return fmt.Errorf("failed to permanently fail job: %w", err)
		}
		logger.Errorf("Job permanently failed: id=%s type=%s attempts=%d error=%s",
			job.ID, job.JobType, job.Attempts, errMsg)
		return nil
	}

	// Reschedule with backoff: delay * 2^(attempts-1)
	delay := retryDelay * time.Duration(1<<uint(job.Attempts-1))
	nextRun := time.Now().Add(delay)
	job.Status = JobStatusQueued
	job.NotBefore = nextRun
	job.LastError = errMsg
	if err := q.store.SaveJob(ctx, job); err != nil {
		return fmt.Errorf("failed to reschedule job: %w", err)
	}

	logger.Warn(fmt.Sprintf("Job rescheduled: id=%s type=%s attempt=%d/%d nextRun=%s error=%s",
		job.ID, job.JobType, job.Attempts, maxAttempts, nextRun.Format(time.RFC3339), errMsg))
	return nil
}
