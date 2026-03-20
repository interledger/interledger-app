package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"html"
	"net/http"
	"strings"
	"time"

	"gitlab.com/fynbos/mock/mockchimoney/internal/models"
	"gitlab.com/fynbos/mock/mockchimoney/internal/storage"
	"gitlab.com/fynbos/mock/mockchimoney/web"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type initiatePaymentRequest struct {
	Amount              string `json:"amount"`
	Currency            string `json:"currency"`
	SubAccount          string `json:"subAccount"`
	PayerEmail          string `json:"payerEmail"`
	RedirectURL         string `json:"redirect_url"`
	TurnOffNotification bool   `json:"turnOffNotification"`
}

type verifyPaymentRequest struct {
	ID         string `json:"id"`
	SubAccount string `json:"subAccount"`
}

func (h *Handler) InitiatePayment(w http.ResponseWriter, r *http.Request) {
	var req initiatePaymentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondErr(w, http.StatusBadRequest, "invalid request body")
		return
	}

	amountStr, ok := requireTrimmedField(req.Amount)
	if !ok {
		h.respondErr(w, http.StatusBadRequest, "amount is required")
		return
	}

	amount, err := parseAmountString(amountStr)
	if err != nil {
		h.respondErr(w, http.StatusBadRequest, "amount is invalid")
		return
	}

	currency, ok := requireTrimmedField(req.Currency)
	if !ok {
		h.respondErr(w, http.StatusBadRequest, "currency is required")
		return
	}
	if _, exists := supportedPaymentCurrencies[currency]; !exists {
		h.respondErr(w, http.StatusBadRequest, "currency is not supported")
		return
	}

	payerEmail, ok := requireTrimmedField(req.PayerEmail)
	if !ok {
		h.respondErr(w, http.StatusBadRequest, "payerEmail is required")
		return
	}

	if err := h.ensureSubAccountExists(r.Context(), req.SubAccount); err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			h.respondErr(w, http.StatusBadRequest, "subAccount not found")
			return
		}
		h.respondErr(w, http.StatusInternalServerError, "failed to validate subAccount")
		return
	}

	issueID := generateIssueID(req.SubAccount)
	chiRef := uuid.NewString()
	payment := models.Payment{
		ID:          uuid.NewString(),
		IssueID:     issueID,
		SubAccount:  strings.TrimSpace(req.SubAccount),
		Amount:      amount,
		Currency:    currency,
		Status:      "pending",
		PayerEmail:  payerEmail,
		RedirectURL: strings.TrimSpace(req.RedirectURL),
		ChiRef:      chiRef,
		CreatedAt:   time.Now().UTC(),
	}

	if _, err := h.store.CreatePayment(r.Context(), payment); err != nil {
		h.respondErr(w, http.StatusInternalServerError, "failed to create payment")
		return
	}

	paymentLink := strings.TrimRight(h.config.PublicBaseURL, "/") + "/pay/" + issueID
	h.respondOK(w, http.StatusOK, map[string]any{
		"paymentLink": paymentLink,
		"issueID":     issueID,
		"chiRef":      chiRef,
		"status":      "pending",
		"payerEmail":  payerEmail,
	})
}

func (h *Handler) VerifyPayment(w http.ResponseWriter, r *http.Request) {
	var req verifyPaymentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondErr(w, http.StatusBadRequest, "invalid request body")
		return
	}

	issueID, ok := requireTrimmedField(req.ID)
	if !ok {
		h.respondErr(w, http.StatusBadRequest, "id is required")
		return
	}

	payment, err := h.store.GetPaymentByIssueID(r.Context(), issueID)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			h.respondErr(w, http.StatusNotFound, "payment not found")
			return
		}
		h.respondErr(w, http.StatusInternalServerError, "failed to verify payment")
		return
	}

	h.respondOK(w, http.StatusOK, map[string]any{
		"id":         payment.ID,
		"issueID":    payment.IssueID,
		"amount":     fmt.Sprintf("%.2f", payment.Amount),
		"currency":   payment.Currency,
		"subAccount": payment.SubAccount,
		"status":     payment.Status,
		"chiRef":     payment.ChiRef,
		"payerEmail": payment.PayerEmail,
		"issueDate":  payment.CreatedAt.Format(time.RFC3339),
		"meta": map[string]any{
			"amount": payment.Amount,
			"processingFee": map[string]any{
				"amount":      h.config.InteracFeeFlat,
				"currency":    payment.Currency,
				"grossAmount": payment.Amount + h.config.InteracFeeFlat,
				"netAmount":   payment.Amount,
				"provider":    "interac",
			},
		},
	})
}

func (h *Handler) PayPage(w http.ResponseWriter, r *http.Request) {
	issueID := strings.TrimSpace(chi.URLParam(r, "issueID"))
	if issueID == "" {
		h.respondErr(w, http.StatusNotFound, "payment not found")
		return
	}

	if _, err := h.store.GetPaymentByIssueID(r.Context(), issueID); err != nil {
		h.respondErr(w, http.StatusNotFound, "payment not found")
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	page := strings.ReplaceAll(web.PayHTML, "{{ISSUE_ID}}", html.EscapeString(issueID))
	_, _ = w.Write([]byte(page))
}

func (h *Handler) ConfirmPayPage(w http.ResponseWriter, r *http.Request) {
	issueID := strings.TrimSpace(chi.URLParam(r, "issueID"))
	if issueID == "" {
		h.respondErr(w, http.StatusNotFound, "payment not found")
		return
	}

	payment, err := h.store.UpdatePaymentStatus(r.Context(), issueID, "redeemed")
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			h.respondErr(w, http.StatusNotFound, "payment not found")
			return
		}
		h.respondErr(w, http.StatusInternalServerError, "failed to update payment")
		return
	}

	baseDelay := time.Duration(h.config.WebhookMinDelaySec * float64(time.Second))
	h.enqueueWebhook(map[string]any{
		"eventType": "charge.interac.completed",
		"issueID":   payment.IssueID,
		"status":    "completed",
	}, baseDelay, nil)
	h.enqueueWebhook(map[string]any{
		"eventType": "chimoney.redeem.completed",
		"issueID":   payment.IssueID,
		"status":    "completed",
	}, baseDelay+50*time.Millisecond, nil)

	redirectTarget := payment.RedirectURL
	if strings.TrimSpace(redirectTarget) == "" {
		redirectTarget = "https://app.test/callbacks/chimoney"
	}
	redirectTarget = appendQuery(redirectTarget, map[string]string{
		"issueID": issueID,
		"status":  "success",
	})

	http.Redirect(w, r, redirectTarget, http.StatusFound)
}
