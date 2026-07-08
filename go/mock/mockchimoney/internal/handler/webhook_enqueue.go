package handler

import (
	"context"
	"time"

	"github.com/interledger/interledger-app/go/mock/mockchimoney/internal/jobs"
	"github.com/interledger/interledger-app/go/mock/mockchimoney/internal/logger"

	"go.uber.org/zap"
)

func (h *Handler) enqueueWebhook(payload map[string]any, delay time.Duration, afterSend func(context.Context) error) {
	if h.queue == nil || h.sender == nil {
		return
	}

	if err := h.queue.Enqueue(jobs.Job{
		Delay: delay,
		Run: func(ctx context.Context) error {
			if err := h.sender.Send(ctx, h.config.WebhookURL, h.config.WebhookSecret, payload); err != nil {
				return err
			}
			if afterSend != nil {
				return afterSend(ctx)
			}
			return nil
		},
	}); err != nil {
		logger.Warn("failed to enqueue webhook", zap.Error(err))
	}
}
