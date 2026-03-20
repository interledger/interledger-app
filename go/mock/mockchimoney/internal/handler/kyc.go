package handler

import (
	"errors"
	"html"
	"net/http"
	"strings"
	"time"

	"gitlab.com/fynbos/mock/mockchimoney/internal/storage"
	"gitlab.com/fynbos/mock/mockchimoney/web"

	"github.com/go-chi/chi/v5"
)

func (h *Handler) KYCPage(w http.ResponseWriter, r *http.Request) {
	externalID := strings.TrimSpace(chi.URLParam(r, "externalID"))
	if externalID == "" {
		h.respondErr(w, http.StatusNotFound, "sub-account not found")
		return
	}

	if _, err := h.store.GetSubAccount(r.Context(), externalID); err != nil {
		h.respondErr(w, http.StatusNotFound, "sub-account not found")
		return
	}

	redirectURL := strings.TrimSpace(r.URL.Query().Get("redirect"))
	if redirectURL == "" {
		h.respondErr(w, http.StatusBadRequest, "redirect is required")
		return
	}

	page := strings.ReplaceAll(web.KYCHTML, "{{EXTERNAL_ID}}", html.EscapeString(externalID))
	page = strings.ReplaceAll(page, "{{REDIRECT_URL}}", html.EscapeString(redirectURL))

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(page))
}

func (h *Handler) KYCApprove(w http.ResponseWriter, r *http.Request) {
	h.handleKYCDecision(w, r, "completed", "user.kyc.completed")
}

func (h *Handler) KYCDecline(w http.ResponseWriter, r *http.Request) {
	h.handleKYCDecision(w, r, "declined", "user.kyc.declined")
}

func (h *Handler) handleKYCDecision(w http.ResponseWriter, r *http.Request, targetStatus string, eventType string) {
	externalID := strings.TrimSpace(chi.URLParam(r, "externalID"))
	if externalID == "" {
		h.respondErr(w, http.StatusNotFound, "sub-account not found")
		return
	}

	account, err := h.store.GetSubAccount(r.Context(), externalID)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			h.respondErr(w, http.StatusNotFound, "sub-account not found")
			return
		}
		h.respondErr(w, http.StatusInternalServerError, "failed to load sub-account")
		return
	}

	if targetStatus == "completed" && account.KYCStatus == "completed" {
		h.respondErr(w, http.StatusConflict, "KYC is already completed")
		return
	}

	if _, err := h.store.UpdateSubAccountKYCStatus(r.Context(), externalID, targetStatus); err != nil {
		h.respondErr(w, http.StatusInternalServerError, "failed to update KYC status")
		return
	}

	redirectURL := strings.TrimSpace(r.FormValue("redirect"))
	if redirectURL == "" {
		redirectURL = strings.TrimSpace(r.URL.Query().Get("redirect"))
	}
	if redirectURL == "" {
		h.respondErr(w, http.StatusBadRequest, "redirect is required")
		return
	}

	h.enqueueWebhook(map[string]any{
		"eventType": eventType,
		"userID":    externalID,
	}, time.Duration(h.config.WebhookMinDelaySec*float64(time.Second)), nil)

	if targetStatus == "declined" {
		redirectURL = appendQuery(redirectURL, map[string]string{"status": "failed"})
	}

	http.Redirect(w, r, redirectURL, http.StatusFound)
}
