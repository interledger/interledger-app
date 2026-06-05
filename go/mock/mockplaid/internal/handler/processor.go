package handler

import (
	"net/http"

	"github.com/google/uuid"
)

// CreateProcessorToken handles POST /processor/token/create. Mints a one-shot
// processor token bound to (access_token, account_id, processor).
func (h *Handler) CreateProcessorToken(w http.ResponseWriter, r *http.Request) {
	h.logCreds(r)

	var req struct {
		AccessToken string `json:"access_token"`
		AccountID   string `json:"account_id"`
		Processor   string `json:"processor"`
	}
	if err := h.decodeJSON(r, &req); err != nil {
		h.sendPlaidError(w, http.StatusBadRequest, "INVALID_INPUT", "INVALID_REQUEST_BODY", err.Error())
		return
	}

	item, err := h.requireItem(w, r, req.AccessToken)
	if err != nil {
		return
	}

	found := false
	for _, a := range item.Accounts {
		if a.AccountID == req.AccountID {
			found = true
			break
		}
	}
	if !found {
		h.sendPlaidError(w, http.StatusBadRequest, "INVALID_INPUT", "INVALID_ACCOUNT_ID", "account_id not found on item")
		return
	}

	h.sendJSON(w, http.StatusOK, map[string]interface{}{
		"processor_token": "processor-sandbox-" + uuid.NewString(),
		"request_id":      requestID(),
	})
}
