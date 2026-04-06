package jobs

import (
	"context"
	"time"

	"gitlab.com/fynbos/mock/mockchimoney/internal/logger"

	"go.uber.org/zap"
)

// StartWorker executes queued jobs until context cancellation.
func StartWorker(ctx context.Context, queue *InMemoryQueue) {
	for {
		select {
		case <-ctx.Done():
			return
		case job := <-queue.Jobs():
			if job.Delay > 0 {
				select {
				case <-ctx.Done():
					return
				case <-time.After(job.Delay):
				}
			}

			if job.Run == nil {
				continue
			}

			if err := job.Run(ctx); err != nil {
				logger.Error("job execution failed", zap.Error(err))
			}
		}
	}
}
