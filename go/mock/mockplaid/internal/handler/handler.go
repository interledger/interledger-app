package handler

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"gitlab.com/fynbos/mock/mockplaid/internal/config"
	"gitlab.com/fynbos/mock/mockplaid/internal/logger"
	"gitlab.com/fynbos/mock/mockplaid/internal/storage"
)

// Handler serves the mock Plaid REST + Link surface.
type Handler struct {
	store  storage.Storage
	config *config.Config
}

// NewHandler creates a new handler.
func NewHandler(store storage.Storage, cfg *config.Config) *Handler {
	return &Handler{store: store, config: cfg}
}

// requestID returns a Plaid-style request id for response envelopes / log correlation.
func requestID() string {
	return "mockplaid-" + uuid.NewString()
}

// sendJSON writes a JSON response.
func (h *Handler) sendJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

// sendPlaidError writes Plaid's error envelope so the backend SDK's wrapPlaidError
// surfaces a meaningful error_code/error_message.
func (h *Handler) sendPlaidError(w http.ResponseWriter, status int, errorType, errorCode, message string) {
	h.sendJSON(w, status, map[string]interface{}{
		"error_type":      errorType,
		"error_code":      errorCode,
		"error_message":   message,
		"display_message": nil,
		"request_id":      requestID(),
	})
}

// logCreds debug-logs the inbound Plaid auth headers WITHOUT logging secret values
// (presence only) — per the Plaid POC logging rule (never log raw tokens/secrets).
func (h *Handler) logCreds(r *http.Request) {
	logger.Debug("plaid request",
		zap.String("method", r.Method),
		zap.String("path", r.URL.Path),
		zap.Bool("client_id_present", r.Header.Get("PLAID-CLIENT-ID") != ""),
		zap.Bool("secret_present", r.Header.Get("PLAID-SECRET") != ""),
	)
}

// NotImplemented is the shared stub for Plaid REST routes not yet built. Returns
// 501 with a Plaid-style envelope so callers can distinguish "route exists, not
// built" from "wrong URL" (404).
func (h *Handler) NotImplemented(w http.ResponseWriter, r *http.Request) {
	h.logCreds(r)
	h.sendPlaidError(w, http.StatusNotImplemented, "INTERNAL", "NOT_IMPLEMENTED", r.URL.Path)
}

// Reset wipes mock state (test convenience).
func (h *Handler) Reset(w http.ResponseWriter, r *http.Request) {
	if err := h.store.Reset(r.Context()); err != nil {
		h.sendPlaidError(w, http.StatusInternalServerError, "INTERNAL", "RESET_FAILED", err.Error())
		return
	}
	h.sendJSON(w, http.StatusOK, map[string]string{"status": "reset"})
}
