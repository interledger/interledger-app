package handler

import (
	"context"
	"fmt"

	"gitlab.com/fynbos/mock/mockpti/internal/jobs"
	"gitlab.com/fynbos/mock/mockpti/internal/models"
)

// NewUserAssessmentWebhookJobHandler returns a worker handler for USER_ASSESSMENT jobs.
func (h *Handler) NewUserAssessmentWebhookJobHandler() jobs.JobHandler {
	return func(ctx context.Context, job *models.Job) error {
		requestID, _ := job.Data["request_id"].(string)
		userID, _ := job.Data["user_id"].(string)
		if userID == "" {
			return fmt.Errorf("missing user_id in webhook job data")
		}

		assessment, err := h.store.GetLatestAssessment(ctx, userID)
		if err != nil {
			return fmt.Errorf("failed to load assessment for user %s: %w", userID, err)
		}
		if requestID != "" && assessment.RequestID != requestID {
			return fmt.Errorf("assessment request id mismatch: want %s got %s", requestID, assessment.RequestID)
		}

		return h.webhook.SendUserAssessment(ctx, assessment)
	}
}

// NewTransactionStatusWebhookJobHandler returns a worker handler for TRANSACTION_STATUS jobs.
func (h *Handler) NewTransactionStatusWebhookJobHandler() jobs.JobHandler {
	return func(ctx context.Context, job *models.Job) error {
		requestID, _ := job.Data["request_id"].(string)
		if requestID == "" {
			return fmt.Errorf("missing request_id in transaction webhook job data")
		}

		tx, err := h.store.GetTransaction(ctx, requestID)
		if err != nil {
			return fmt.Errorf("failed to load transaction %s: %w", requestID, err)
		}

		return h.webhook.SendTransactionStatus(ctx, tx)
	}
}
