package handler

import (
	"encoding/json"
	"net/http"
	"strings"
)

type feeEstimateRequest struct {
	Amount    json.RawMessage `json:"amount"`
	Currency  string          `json:"currency"`
	Rail      string          `json:"rail"`
	Direction string          `json:"direction"`
}

func (h *Handler) FeeEstimate(w http.ResponseWriter, r *http.Request) {
	var req feeEstimateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondErr(w, http.StatusBadRequest, "invalid request body")
		return
	}

	amount, err := parseFlexibleFloat(req.Amount)
	if err != nil || amount <= 0 {
		h.respondErr(w, http.StatusBadRequest, "amount is required")
		return
	}

	rail := strings.TrimSpace(req.Rail)
	currency := strings.TrimSpace(req.Currency)
	if rail == "" && !strings.EqualFold(currency, "USD") {
		h.respondErr(w, http.StatusBadRequest, "currency must be USD when rail is not specified")
		return
	}

	direction := strings.TrimSpace(req.Direction)
	if direction == "" {
		direction = "payout"
	}

	totalFee := h.config.InteracFeeFlat
	netAmount := amount - totalFee

	h.respondOK(w, http.StatusOK, map[string]any{
		"amount":    amount,
		"currency":  currency,
		"rail":      rail,
		"direction": direction,
		"totalFee":  totalFee,
		"netAmount": netAmount,
	})
}
