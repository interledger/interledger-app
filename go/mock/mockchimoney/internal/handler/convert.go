package handler

import (
	"net/http"
	"strconv"
	"strings"
	"time"
)

func (h *Handler) ConvertLocalAmountToUSD(w http.ResponseWriter, r *http.Request) {
	originCurrency, ok := requireTrimmedField(r.URL.Query().Get("originCurrency"))
	if !ok {
		h.respondErr(w, http.StatusBadRequest, "originCurrency is required")
		return
	}

	amountRaw, ok := requireTrimmedField(r.URL.Query().Get("amountInOriginCurrency"))
	if !ok {
		h.respondErr(w, http.StatusBadRequest, "amountInOriginCurrency is required")
		return
	}

	amount, err := strconv.ParseFloat(amountRaw, 64)
	if err != nil {
		h.respondErr(w, http.StatusBadRequest, "amountInOriginCurrency is invalid")
		return
	}

	amountInUSD := amount * h.config.CADToUSDRate
	expiresAt := time.Now().UTC().Add(5 * time.Minute)

	h.respondOK(w, http.StatusOK, map[string]any{
		"originCurrency":         strings.ToUpper(originCurrency),
		"amountInOriginCurrency": amountRaw,
		"amountInUSD":            amountInUSD,
		"validUntil":             expiresAt.Format(time.RFC3339),
		"expiresAt":              expiresAt.Format(time.RFC3339),
		"expiresAtTimestamp":     expiresAt.Unix(),
	})
}
