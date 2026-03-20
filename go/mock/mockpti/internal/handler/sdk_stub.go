package handler

import (
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"gitlab.com/fynbos/mock/mockpti/internal/jobs"
	"gitlab.com/fynbos/mock/mockpti/internal/models"
	"gitlab.com/fynbos/mock/mockpti/internal/storage"
)

//go:embed web/forms.html
var formsHTML string

const mockPTISDKScript = `(function () {
  window.PTI = {
    init: function (config) {
      window.__mockPtiInit = config || {};
    },
    form: function (options) {
      options = options || {};

      var formsBase = (window.__mockPtiInit && window.__mockPtiInit.ptiFormsUrl) ||
        'https://mockpti.interledger.test/forms';
      var params = new URLSearchParams({
        type: options.type || 'KYC',
        requestId: options.requestId || '',
        userId: options.userId || '',
        scenarioId: options.scenarioId || '',
        auto: options.type === 'ADD_CC' ? '1' : '0'
      });

      if (options.parentElement && formsBase) {
        var iframe = document.createElement('iframe');
        iframe.setAttribute('title', 'Mock PTI Form');
        iframe.setAttribute('src', formsBase + '?' + params.toString());
        iframe.setAttribute('style', 'width:100%;height:100%;min-height:720px;border:0;');
        options.parentElement.innerHTML = '';
        options.parentElement.appendChild(iframe);
      }
    }
  };
})();
`

func (h *Handler) SDKScript(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/javascript")
	_, _ = fmt.Fprint(w, mockPTISDKScript)
}

func (h *Handler) FormsLanding(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = fmt.Fprint(w, formsHTML)
}

type completeAssessmentRequest struct {
	UserID      string `json:"userId"`
	RequestID   string `json:"requestId"`
	DateOfBirth string `json:"dateOfBirth"`
}

// CompleteAssessmentFromForm marks a user assessment as accepted and schedules webhook delivery.
// This endpoint is public and intended for local iframe simulation only.
func (h *Handler) CompleteAssessmentFromForm(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.sendError(w, http.StatusMethodNotAllowed, "method_not_allowed", "only POST is supported")
		return
	}

	var req completeAssessmentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.sendError(w, http.StatusBadRequest, "invalid_request", "invalid request body")
		return
	}
	if req.UserID == "" {
		h.sendError(w, http.StatusBadRequest, "invalid_request", "userId is required")
		return
	}
	if req.RequestID == "" {
		h.sendError(w, http.StatusBadRequest, "invalid_request", "requestId is required")
		return
	}

	user, err := h.store.GetUser(r.Context(), req.UserID)
	if err != nil {
		if errors.Is(err, storage.ErrUserNotFound) {
			h.sendError(w, http.StatusNotFound, "not_found", "user not found")
			return
		}
		h.sendError(w, http.StatusInternalServerError, "internal_error", "failed to load user")
		return
	}

	if req.DateOfBirth != "" {
		user.DateOfBirth = req.DateOfBirth
	} else if user.DateOfBirth == "" {
		h.sendError(w, http.StatusBadRequest, "invalid_request", "dateOfBirth is required")
		return
	}

	if err := h.store.UpdateUser(r.Context(), user); err != nil {
		h.sendError(w, http.StatusInternalServerError, "internal_error", "failed to update user")
		return
	}

	assessment := &models.Assessment{
		ResourceType: "assessment",
		ClientID:     h.config.ClientID,
		RequestID:    req.RequestID,
		UserID:       req.UserID,
		Date:         time.Now().Format(time.RFC3339),
		Assessment:   "ACCEPTED",
		Tier:         1,
	}

	if err := h.store.SaveAssessment(r.Context(), assessment); err != nil {
		h.sendError(w, http.StatusInternalServerError, "internal_error", "failed to save assessment")
		return
	}

	h.enqueueWebhook(jobs.JobTypeUserAssessmentWebhook, map[string]interface{}{
		"user_id":    req.UserID,
		"request_id": req.RequestID,
	})

	h.sendJSON(w, http.StatusOK, map[string]string{"status": "accepted"})
}
