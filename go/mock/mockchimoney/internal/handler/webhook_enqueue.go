package handler

import (
	"context"
	"time"

	"gitlab.com/fynbos/mock/mockchimoney/internal/jobs"
)

func (h *Handler) enqueueWebhook(payload map[string]any, delay time.Duration, afterSend func(context.Context) error) {
	if h.queue == nil || h.sender == nil {
		return
	}

	_ = h.queue.Enqueue(jobs.Job{
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
	})
}
