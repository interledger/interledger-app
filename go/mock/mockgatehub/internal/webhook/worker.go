package webhook

import (
	"context"
	"fmt"
	"time"

	"github.com/interledger/interledger-app/go/mock/mockgatehub/internal/logger"

	"go.uber.org/zap"
)

// Worker processes webhook jobs from the queue
type Worker struct {
	queue   *Queue
	manager *Manager
	ctx     context.Context
	cancel  context.CancelFunc
}

// NewWorker creates a new webhook worker
func NewWorker(queue *Queue, manager *Manager) *Worker {
	ctx, cancel := context.WithCancel(context.Background())
	return &Worker{
		queue:   queue,
		manager: manager,
		ctx:     ctx,
		cancel:  cancel,
	}
}

// Start begins processing webhook jobs (blocking)
func (w *Worker) Start() {
	logger.Info("starting webhook worker", zap.Duration("poll_interval", pollInterval))

	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-w.ctx.Done():
			logger.Info("stopping webhook worker")
			return
		case <-ticker.C:
			w.processReadyJobs()
		}
	}
}

// Stop gracefully stops the worker
func (w *Worker) Stop() {
	w.cancel()
}

// processReadyJobs fetches and processes all ready jobs sequentially
func (w *Worker) processReadyJobs() {
	jobs, err := w.queue.GetReadyJobs(w.ctx, batchSize)
	if err != nil {
		logger.Error("failed to fetch ready jobs", zap.Error(err))
		return
	}

	if len(jobs) == 0 {
		return // No jobs ready
	}

	logger.Info("found ready jobs to process", zap.Int("count", len(jobs)))

	for _, job := range jobs {
		w.processJob(job)
	}
}

// processJob processes a single webhook job
func (w *Worker) processJob(job *Job) {
	logger.Info("processing webhook job",
		zap.String("job_id", job.ID),
		zap.String("event", job.EventType),
		zap.String("user", job.UserID),
		zap.Int("attempt", job.Attempts+1),
		zap.Int("max_attempts", maxAttempts),
	)

	// Send webhook using manager's send method
	err := w.manager.send(job.EventType, job.UserID, job.Data)

	if err != nil {
		// Mark as failed and reschedule
		errMsg := err.Error()
		if markErr := w.queue.MarkFailed(w.ctx, job.ID, errMsg); markErr != nil {
			logger.Error("failed to mark job as failed", zap.Error(markErr))
		}
		logger.Warn("webhook job failed", zap.String("job_id", job.ID), zap.String("error", errMsg))
		return
	}

	// Mark as completed
	if err := w.queue.MarkCompleted(w.ctx, job.ID); err != nil {
		logger.Error("failed to mark job as completed", zap.Error(err))
		return
	}

	logger.Info("webhook job completed successfully", zap.String("job_id", job.ID))
}

// StartAsync starts the worker in a goroutine
func (w *Worker) StartAsync() {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				logger.Error("panic in webhook worker", zap.Any("panic", r))
				// Restart worker after panic, but only if context is still active
				time.Sleep(5 * time.Second)
				select {
				case <-w.ctx.Done():
					logger.Info("not restarting webhook worker after panic because context is canceled")
				default:
					logger.Info("restarting webhook worker after panic")
					w.StartAsync()
				}
			}
		}()
		w.Start()
	}()
}

// GetStats returns queue statistics (for debugging/monitoring)
func (w *Worker) GetStats(ctx context.Context) (map[string]int64, error) {
	// Count total jobs in queue
	totalInQueue, err := w.queue.client.ZCard(ctx, queueKey).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to count queue size: %w", err)
	}

	// Count ready jobs
	now := time.Now().Unix()
	readyCount, err := w.queue.client.ZCount(ctx, queueKey, "-inf", fmt.Sprintf("%d", now)).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to count ready jobs: %w", err)
	}

	return map[string]int64{
		"total_in_queue": totalInQueue,
		"ready_now":      readyCount,
		"scheduled":      totalInQueue - readyCount,
	}, nil
}
