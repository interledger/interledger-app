package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"

	"gitlab.com/fynbos/mock/mockpti/internal/logger"
	"gitlab.com/fynbos/mock/mockpti/internal/models"
	"gitlab.com/fynbos/mock/mockpti/internal/storage"
	"gitlab.com/fynbos/mock/mockpti/internal/utils"
)

// StartUserAssessment handles POST /users/assessments.
func (h *Handler) StartUserAssessment(w http.ResponseWriter, r *http.Request) {
	var req models.StartAssessmentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.sendError(w, http.StatusBadRequest, "invalid_request", "invalid request body")
		return
	}

	if req.ID == "" {
		h.sendError(w, http.StatusBadRequest, "invalid_request", "user id is required")
		return
	}

	// Verify user exists
	_, err := h.store.GetUser(r.Context(), req.ID)
	if err != nil {
		if errors.Is(err, storage.ErrUserNotFound) {
			h.sendError(w, http.StatusNotFound, "not_found", "user not found")
			return
		}
		h.sendError(w, http.StatusInternalServerError, "internal_error", "failed to get user")
		return
	}

	requestID := r.Header.Get("x-pti-request-id")
	if requestID == "" {
		requestID = utils.GenerateUUID()
	}

	assessment := &models.Assessment{
		ResourceType: "assessment",
		ClientID:     h.config.ClientID,
		RequestID:    requestID,
		UserID:       req.ID,
		Date:         time.Now().Format(time.RFC3339),
		Assessment:   "approved",
		Tier:         1,
	}

	if err := h.store.SaveAssessment(r.Context(), assessment); err != nil {
		logger.Errorf("failed to save assessment: %v", err)
		h.sendError(w, http.StatusInternalServerError, "internal_error", "failed to start assessment")
		return
	}

	logger.Infof("Started assessment for user %s, requestID=%s", req.ID, requestID)

	h.sendJSON(w, http.StatusOK, models.IDResponse{
		ID:   requestID,
		Link: "/users/" + req.ID + "/assessments",
	})
}

// GetUserAssessment handles GET /users/{id}/assessments.
func (h *Handler) GetUserAssessment(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "id")
	if userID == "" {
		h.sendError(w, http.StatusBadRequest, "invalid_request", "user id is required")
		return
	}

	assessment, err := h.store.GetLatestAssessment(r.Context(), userID)
	if err != nil {
		if errors.Is(err, storage.ErrAssessmentNotFound) {
			h.sendError(w, http.StatusNotFound, "not_found", "assessment not found")
			return
		}
		h.sendError(w, http.StatusInternalServerError, "internal_error", "failed to get assessment")
		return
	}

	h.sendJSON(w, http.StatusOK, assessment)
}
