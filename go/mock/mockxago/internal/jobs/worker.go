package jobs

import (
	"context"
	"fmt"
	"time"

	"gitlab.com/fynbos/mockxago/internal/logger"
	"gitlab.com/fynbos/mockxago/internal/models"
)

// JobHandler is a function that processes a job. Return nil on success, error on failure.
type JobHandler func(ctx context.Context, job *models.Job) error

// Worker polls the queue and processes jobs sequentially.
type Worker struct {
	queue    *Queue
	handlers map[string]JobHandler
	ctx      context.Context
	cancel   context.CancelFunc
}

// NewWorker creates a new worker attached to the given queue.
func NewWorker(queue *Queue) *Worker {
	ctx, cancel := context.WithCancel(context.Background())
	return &Worker{
		queue:    queue,
		handlers: make(map[string]JobHandler),
		ctx:      ctx,
		cancel:   cancel,
	}
}

// RegisterHandler registers a handler for a given job type.
func (w *Worker) RegisterHandler(jobType string, handler JobHandler) {
	w.handlers[jobType] = handler
}

// Start begins polling the queue (blocking).
func (w *Worker) Start() {
	logger.Infof("Job worker started (poll_interval=%s, batch_size=%d, max_attempts=%d)",
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

// Stop gracefully stops the worker.
func (w *Worker) Stop() {
	w.cancel()
}

// StartAsync starts the worker in a background goroutine with panic recovery.
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

// processReadyJobs fetches and processes all ready jobs sequentially.
func (w *Worker) processReadyJobs() {
	jobs, err := w.queue.GetReadyJobs(batchSize)
	if err != nil {
		logger.Errorf("Failed to get ready jobs: %v", err)
		return
	}
	if len(jobs) == 0 {
		return
	}

	logger.Infof("Processing %d ready job(s)", len(jobs))

	for _, job := range jobs {
		w.processJob(job)
	}
}

// processJob processes a single job by dispatching to the registered handler.
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

	err := handler(w.ctx, job)
	if err != nil {
		_ = w.queue.MarkFailed(job.ID, err.Error())
		logger.Warnf("Job failed: id=%s type=%s error=%s", job.ID, job.JobType, err.Error())
		return
	}

	_ = w.queue.MarkCompleted(job.ID)
}
