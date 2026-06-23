package jobs

import (
	"context"
	"fmt"
	"time"

	"github.com/interledger/interledger-app/go/mock/mockpti/internal/logger"
	"github.com/interledger/interledger-app/go/mock/mockpti/internal/models"
)

// JobHandler is a function that processes a single job.
// Return nil on success, non-nil error to trigger retry/failure logic.
type JobHandler func(ctx context.Context, job *models.Job) error

// Worker polls the queue and dispatches jobs to registered handlers.
type Worker struct {
	queue    *Queue
	handlers map[string]JobHandler
	ctx      context.Context
	cancel   context.CancelFunc
}

// NewWorker creates a new worker backed by the given queue.
func NewWorker(queue *Queue) *Worker {
	ctx, cancel := context.WithCancel(context.Background())
	return &Worker{
		queue:    queue,
		handlers: make(map[string]JobHandler),
		ctx:      ctx,
		cancel:   cancel,
	}
}

// RegisterHandler registers a handler for the given job type.
func (w *Worker) RegisterHandler(jobType string, handler JobHandler) {
	w.handlers[jobType] = handler
}

// Start polls the queue and processes jobs until the worker is stopped.
// This call blocks; use StartAsync for background operation.
func (w *Worker) Start() {
	logger.Infof("Job worker started (poll_interval=%s, batch=%d, max_attempts=%d)",
		pollInterval, batchSize, maxAttempts)

	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-w.ctx.Done():
			logger.Infof("Job worker stopping")
			return
		case <-ticker.C:
			w.processReadyJobs()
		}
	}
}

// Stop signals the worker to stop after the current batch.
func (w *Worker) Stop() {
	w.cancel()
}

// StartAsync launches the worker in a background goroutine with panic recovery.
func (w *Worker) StartAsync() {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				logger.Errorf("Panic in job worker: %v", r)
				time.Sleep(5 * time.Second)
				logger.Infof("Restarting job worker after panic")
				w.StartAsync()
			}
		}()
		w.Start()
	}()
}

func (w *Worker) processReadyJobs() {
	jobs, err := w.queue.GetReadyJobs(batchSize)
	if err != nil {
		logger.Errorf("Failed to get ready jobs: %v", err)
		return
	}
	for _, job := range jobs {
		w.processJob(job)
	}
}

func (w *Worker) processJob(job *models.Job) {
	handler, ok := w.handlers[job.JobType]
	if !ok {
		errMsg := fmt.Sprintf("no handler registered for job type: %s", job.JobType)
		logger.Errorf("Job %s: %s", job.ID, errMsg)
		_ = w.queue.MarkFailed(job.ID, errMsg)
		return
	}

	logger.Infof("Processing job: id=%s type=%s attempt=%d/%d",
		job.ID, job.JobType, job.Attempts+1, maxAttempts)

	if err := w.queue.MarkProcessing(job.ID); err != nil {
		logger.Errorf("failed to mark job processing: id=%s type=%s error=%v", job.ID, job.JobType, err)
		return
	}

	if err := handler(w.ctx, job); err != nil {
		if markErr := w.queue.MarkFailed(job.ID, err.Error()); markErr != nil {
			logger.Errorf("failed to mark job failed: id=%s type=%s error=%v", job.ID, job.JobType, markErr)
			return
		}
		logger.Warn(fmt.Sprintf("Job failed: id=%s type=%s error=%s", job.ID, job.JobType, err.Error()))
		return
	}

	if err := w.queue.MarkDelivered(job.ID); err != nil {
		logger.Errorf("failed to mark job delivered: id=%s type=%s error=%v", job.ID, job.JobType, err)
	}
}
