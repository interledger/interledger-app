package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"gitlab.com/fynbos/mock/mockchimoney/internal/storage"
)

type transferRequest struct {
	SubAccount          string `json:"subAccount"`
	Receiver            string `json:"receiver"`
	AmountToSend        string `json:"amountToSend"`
	OriginCurrency      string `json:"originCurrency"`
	DestinationCurrency string `json:"destinationCurrency"`
	TurnOffNotification bool   `json:"turnOffNotification"`
	SendViaInterledger  bool   `json:"sendViaInterledger"`
}

// Transfer moves funds between two Chimoney wallets.
func (h *Handler) Transfer(w http.ResponseWriter, r *http.Request) {
	var req transferRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondErr(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if _, ok := requireTrimmedField(req.AmountToSend); !ok {
		h.respondErr(w, http.StatusBadRequest, "amountToSend is required")
		return
	}

	if _, ok := requireTrimmedField(req.OriginCurrency); !ok {
		h.respondErr(w, http.StatusBadRequest, "originCurrency is required")
		return
	}

	if _, ok := requireTrimmedField(req.DestinationCurrency); !ok {
		h.respondErr(w, http.StatusBadRequest, "destinationCurrency is required")
		return
	}

	if err := h.ensureSubAccountExists(r.Context(), req.SubAccount); err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			h.respondErr(w, http.StatusBadRequest, "sender subAccount not found")
			return
		}

		h.respondErr(w, http.StatusInternalServerError, "failed to validate sender subAccount")
		return
	}

	// sendViaInterledger is intentionally accepted and ignored in the mock.
	h.respondOK(w, http.StatusOK, map[string]any{})
}
