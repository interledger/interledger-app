package handler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"gitlab.com/fynbos/mock/mockchimoney/internal/models"
	"gitlab.com/fynbos/mock/mockchimoney/internal/storage"

	"github.com/google/uuid"
)

type interacItem struct {
	Name      string          `json:"name"`
	Email     string          `json:"email"`
	Amount    json.RawMessage `json:"amount"`
	Narration string          `json:"narration"`
}

type payoutInteracRequest struct {
	Interacs            []interacItem `json:"interacs"`
	DebitCurrency       string        `json:"debitCurrency"`
	SubAccount          string        `json:"subAccount"`
	TurnOffNotification bool          `json:"turnOffNotification"`
}

type payoutStatusRequest struct {
	ChiRef              string `json:"chiRef"`
	SubAccount          string `json:"subAccount"`
	TurnOffNotification bool   `json:"turnOffNotification"`
}

func (h *Handler) PayoutInterac(w http.ResponseWriter, r *http.Request) {
	var req payoutInteracRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondErr(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if _, ok := requireTrimmedField(req.DebitCurrency); !ok {
		h.respondErr(w, http.StatusBadRequest, "debitCurrency is required")
		return
	}
	if len(req.Interacs) == 0 {
		h.respondErr(w, http.StatusBadRequest, "interacs is required")
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

	items := make([]map[string]any, 0, len(req.Interacs))
	baseDelay := time.Duration(h.config.WebhookMinDelaySec * float64(time.Second))
	for idx, item := range req.Interacs {
		amount, err := parseFlexibleFloat(item.Amount)
		if err != nil {
			h.respondErr(w, http.StatusBadRequest, "interacs amount is invalid")
			return
		}

		issueID := generateIssueID(req.SubAccount)
		payout := models.Payout{
			ID:           uuid.NewString(),
			IssueID:      issueID,
			SubAccount:   strings.TrimSpace(req.SubAccount),
			Amount:       amount,
			Fee:          h.config.InteracFeeFlat,
			Currency:     strings.TrimSpace(req.DebitCurrency),
			Status:       "pending",
			ChiRef:       uuid.NewString(),
			InteracEmail: strings.TrimSpace(item.Email),
			CreatedAt:    time.Now().UTC(),
		}

		if _, err := h.store.CreatePayout(r.Context(), payout); err != nil {
			h.respondErr(w, http.StatusInternalServerError, "failed to create payout")
			return
		}

		delay := baseDelay + (time.Duration(idx) * 50 * time.Millisecond)
		payload := map[string]any{
			"eventType": "payout.interac.completed",
			"issueID":   payout.IssueID,
			"status":    "completed",
			"amount":    fmt.Sprintf("%.2f", payout.Amount),
			"meta": map[string]any{
				"issuer":      payout.SubAccount,
				"amount":      payout.Amount,
				"currency":    payout.Currency,
				"paymentType": "interac",
			},
		}
		h.enqueueWebhook(payload, delay, func(ctx context.Context) error {
			_, err := h.store.UpdatePayoutStatus(ctx, payout.IssueID, "completed")
			return err
		})

		items = append(items, map[string]any{
			"id":            payout.ID,
			"issueID":       payout.IssueID,
			"amount":        payout.Amount,
			"fee":           payout.Fee,
			"debitCurrency": payout.Currency,
			"type":          "interac",
			"chiref":        payout.ChiRef,
		})
	}

	h.respondOK(w, http.StatusOK, map[string]any{"data": items})
}

func (h *Handler) PayoutStatus(w http.ResponseWriter, r *http.Request) {
	var req payoutStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondErr(w, http.StatusBadRequest, "invalid request body")
		return
	}

	chiRef, ok := requireTrimmedField(req.ChiRef)
	if !ok {
		h.respondErr(w, http.StatusBadRequest, "chiRef is required")
		return
	}

	payout, err := h.store.GetPayoutByChiRef(r.Context(), chiRef)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			h.respondErr(w, http.StatusNotFound, "payout not found")
			return
		}
		h.respondErr(w, http.StatusInternalServerError, "failed to get payout")
		return
	}

	h.respondOK(w, http.StatusOK, map[string]any{
		"id":      payout.ID,
		"amount":  payout.Amount,
		"fee":     payout.Fee,
		"type":    "interac",
		"issueID": payout.IssueID,
		"status":  payout.Status,
	})
}
