package handler

import (
	"encoding/json"
	"net/http"

	"gitlab.com/fynbos/mock/mockpti/internal/config"
	"gitlab.com/fynbos/mock/mockpti/internal/models"
	"gitlab.com/fynbos/mock/mockpti/internal/storage"
)

// Handler handles HTTP requests for the mock PTI service.
type Handler struct {
	store  storage.Storage
	config *config.Config
}

// NewHandler creates a new handler.
func NewHandler(store storage.Storage, cfg *config.Config) *Handler {
	return &Handler{
		store:  store,
		config: cfg,
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
