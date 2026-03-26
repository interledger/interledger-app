package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"gitlab.com/fynbos/mock/mockpti/internal/config"
	"gitlab.com/fynbos/mock/mockpti/internal/jobs"
	"gitlab.com/fynbos/mock/mockpti/internal/logger"
	"gitlab.com/fynbos/mock/mockpti/internal/models"
	"gitlab.com/fynbos/mock/mockpti/internal/storage"
	"gitlab.com/fynbos/mock/mockpti/internal/webhooks"
)

// Handler handles HTTP requests for the mock PTI service.
type Handler struct {
	store   storage.Storage
	config  *config.Config
	queue   *jobs.Queue // nil means webhook jobs are skipped
	webhook *webhooks.Sender
}

// NewHandler creates a new handler without a job queue.
// Call SetQueue to enable async webhook delivery.
func NewHandler(store storage.Storage, cfg *config.Config) *Handler {
	sender := webhooks.NewSender(cfg.WebhookURL)
	if err := sender.ConfigureSecurity(cfg.WebhookSigningJWK, cfg.WebhookEncryptionJWK); err != nil {
		logger.Warn(fmt.Sprintf("Webhook crypto configuration disabled: %v", err))
	}

	return &Handler{
		store:   store,
		config:  cfg,
		webhook: sender,
	}
}

// SetQueue attaches a job queue, enabling async webhook delivery.
func (h *Handler) SetQueue(q *jobs.Queue) {
	h.queue = q
}

// enqueueWebhook safely enqueues a webhook job when a queue is wired, otherwise no-ops.
func (h *Handler) enqueueWebhook(jobType string, data map[string]interface{}) {
	if h.queue == nil {
		return
	}
	if _, err := h.queue.Enqueue(jobType, data, time.Now()); err != nil {
		logger.Warn(fmt.Sprintf("Failed to enqueue %s webhook job: %v", jobType, err))
	}
}

// AuthMiddleware validates the x-pti-client-id header.
func (h *Handler) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		clientID := r.Header.Get("x-pti-client-id")
		if clientID == "" {
			h.sendError(w, http.StatusUnauthorized, "unauthorized", "missing x-pti-client-id header")
			return
		}

		if h.config != nil && h.config.ClientID != "" && clientID != h.config.ClientID {
			h.sendError(w, http.StatusUnauthorized, "unauthorized", "invalid x-pti-client-id")
			return
		}
		next.ServeHTTP(w, r)
	})
}

// sendJSON sends a JSON response.
func (h *Handler) sendJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

// sendError sends an error response.
func (h *Handler) sendError(w http.ResponseWriter, status int, errCode, message string) {
	h.sendJSON(w, status, models.ErrorResponse{
		Error:   errCode,
		Message: message,
	})
}
