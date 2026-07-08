package handler

import (
	"net/http"

	"github.com/interledger/interledger-app/go/mock/mockpti/internal/logger"
)

// Reset handles POST /test/reset — clears all stored data for scenario isolation.
func (h *Handler) Reset(w http.ResponseWriter, r *http.Request) {
	if err := h.store.Reset(r.Context()); err != nil {
		logger.Errorf("failed to reset store: %v", err)
		h.sendError(w, http.StatusInternalServerError, "internal_error", "reset failed")
		return
	}
	logger.Infof("Store reset via /test/reset")
	h.sendJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
