package webhook

import (
	"context"
	"fmt"
	"time"

	"gitlab.com/fynbos/mock/mockgatehub/internal/logger"
	"gitlab.com/fynbos/mock/mockgatehub/internal/utils"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

const (
	queueKey     = "webhooks:queue" // Sorted set: jobID -> not_before timestamp
	jobKeyPrefix = "webhooks:job:"  // Hash per job: webhooks:job:{jobID}
	maxAttempts  = 10               // Give up after 10 failed attempts
	retryDelay   = 30 * time.Second // Fixed 30-second backoff
	pollInterval = 5 * time.Second  // How often worker polls for ready jobs
	batchSize    = 10               // Max jobs to fetch in one poll
)

// Queue manages webhook job persistence in Redis
type Queue struct {
	client   *redis.Client
	minDelay time.Duration
}

// NewQueue creates a new webhook queue.
// The minDelaySec parameter specifies the minimum delay (in seconds) before
// enqueued webhook jobs become eligible for delivery. Values less than 0
// are invalid and should be validated by the caller (enforced in config.Load).
func NewQueue(client *redis.Client, minDelaySec float64) *Queue {
	// Defensive validation: clamp to 0 if accidental negative value slips through
	if minDelaySec < 0 {
		minDelaySec = 0
	}
	return &Queue{
		client:   client,
		minDelay: time.Duration(minDelaySec * float64(time.Second)),
	}
}

// Enqueue adds a new webhook job to the queue.
// offsetDelaySeconds adds extra seconds on top of the queue's minimum delay.
func (q *Queue) Enqueue(ctx context.Context, eventType, userID string, data any, offsetDelaySeconds float64) (string, error) {
	jobID := utils.GenerateUUID()

	// Convert data to map
	dataMap := coerceToMap(data)

	offset := time.Duration(offsetDelaySeconds * float64(time.Second))

	job := &Job{
		ID:        jobID,
		EventType: eventType,
		UserID:    userID,
		Data:      dataMap,
		Attempts:  0,
		Status:    JobStatusPending,
		CreatedAt: time.Now(),
		NotBefore: time.Now().Add(q.minDelay + offset), // Apply minimum delay + offset
	}

	// Serialize job to JSON
	jobJSON, err := job.ToJSON()
	if err != nil {
		return "", fmt.Errorf("failed to serialize job: %w", err)
	}

	// Store job data in hash
	jobKey := jobKeyPrefix + jobID
	err = q.client.Set(ctx, jobKey, jobJSON, 0).Err()
	if err != nil {
		return "", fmt.Errorf("failed to store job: %w", err)
	}

	// Add job to sorted set (score = not_before timestamp)
	score := float64(job.NotBefore.Unix())
	err = q.client.ZAdd(ctx, queueKey, redis.Z{
		Score:  score,
		Member: jobID,
	}).Err()
	if err != nil {
		return "", fmt.Errorf("failed to enqueue job: %w", err)
	}

	logger.Info("enqueued webhook job", zap.String("job_id", jobID), zap.String("event", eventType), zap.String("user", userID))
	return jobID, nil
}

// GetReadyJobs fetches up to 'limit' jobs that are ready to process (not_before <= now)
func (q *Queue) GetReadyJobs(ctx context.Context, limit int64) ([]*Job, error) {
	now := time.Now().Unix()

	// Query sorted set for jobs with score <= now
	results, err := q.client.ZRangeByScore(ctx, queueKey, &redis.ZRangeBy{
		Min:   "-inf",
		Max:   fmt.Sprintf("%d", now),
		Count: limit,
	}).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to query ready jobs: %w", err)
	}

	jobs := make([]*Job, 0, len(results))
	for _, jobID := range results {
		job, err := q.getJob(ctx, jobID)
		if err != nil {
			logger.Error("failed to load job", zap.String("job_id", jobID), zap.Error(err))
			continue
		}

		// Only return pending jobs (ignore completed/failed)
		if job.Status == JobStatusPending {
			jobs = append(jobs, job)
		}
	}

	return jobs, nil
}

// MarkCompleted marks a job as successfully completed
func (q *Queue) MarkCompleted(ctx context.Context, jobID string) error {
	job, err := q.getJob(ctx, jobID)
	if err != nil {
		return err
	}

	now := time.Now()
	job.Status = JobStatusCompleted
	job.CompletedAt = &now

	if err := q.saveJob(ctx, job); err != nil {
		return err
	}

	// Remove from queue (but keep job data for troubleshooting)
	if err := q.client.ZRem(ctx, queueKey, jobID).Err(); err != nil {
		return fmt.Errorf("failed to remove job from queue: %w", err)
	}

	logger.Info("marked job as completed", zap.String("job_id", jobID))
	return nil
}

// MarkFailed marks a job as failed and optionally reschedules it
func (q *Queue) MarkFailed(ctx context.Context, jobID string, errMsg string) error {
	job, err := q.getJob(ctx, jobID)
	if err != nil {
		return err
	}

	job.Attempts++
	job.LastError = errMsg

	// Check if we've exceeded max attempts
	if job.Attempts >= maxAttempts {
		now := time.Now()
		job.Status = JobStatusFailed
		job.CompletedAt = &now

		if err := q.saveJob(ctx, job); err != nil {
			return err
		}

		// Remove from queue (keep job data for troubleshooting)
		if err := q.client.ZRem(ctx, queueKey, jobID).Err(); err != nil {
			return fmt.Errorf("failed to remove failed job from queue: %w", err)
		}

		logger.Error("job permanently failed",
			zap.Int("attempts", job.Attempts),
			zap.String("job_id", jobID),
			zap.String("error", errMsg),
		)
		return nil
	}

	// Reschedule with 30-second delay
	job.NotBefore = time.Now().Add(retryDelay)
	if err := q.saveJob(ctx, job); err != nil {
		return err
	}

	// Update score in sorted set
	score := float64(job.NotBefore.Unix())
	if err := q.client.ZAdd(ctx, queueKey, redis.Z{
		Score:  score,
		Member: jobID,
	}).Err(); err != nil {
		return fmt.Errorf("failed to reschedule job: %w", err)
	}

	logger.Warn("rescheduled job",
		zap.Int("attempt", job.Attempts),
		zap.Int("max_attempts", maxAttempts),
		zap.String("job_id", jobID),
		zap.String("retry_at", job.NotBefore.Format(time.RFC3339)),
	)
	return nil
}

// getJob loads a job from Redis
func (q *Queue) getJob(ctx context.Context, jobID string) (*Job, error) {
	jobKey := jobKeyPrefix + jobID
	data, err := q.client.Get(ctx, jobKey).Bytes()
	if err != nil {
		return nil, fmt.Errorf("failed to load job %s: %w", jobID, err)
	}

	job, err := FromJSON(data)
	if err != nil {
		return nil, fmt.Errorf("failed to deserialize job %s: %w", jobID, err)
	}

	return job, nil
}

// saveJob persists a job to Redis
func (q *Queue) saveJob(ctx context.Context, job *Job) error {
	jobKey := jobKeyPrefix + job.ID
	jobJSON, err := job.ToJSON()
	if err != nil {
		return fmt.Errorf("failed to serialize job: %w", err)
	}

	err = q.client.Set(ctx, jobKey, jobJSON, 0).Err()
	if err != nil {
		return fmt.Errorf("failed to save job: %w", err)
	}

	return nil
}
