package handler

import (
	"net/http"

	"gitlab.com/fynbos/mock/mockplaid/internal/models"
	"gitlab.com/fynbos/mock/mockplaid/internal/storage"
)

type accessTokenRequest struct {
	AccessToken string `json:"access_token"`
}

// ExchangePublicToken handles POST /item/public_token/exchange. Resolves the item
// stored at select time and returns its access_token + item_id.
func (h *Handler) ExchangePublicToken(w http.ResponseWriter, r *http.Request) {
	h.logCreds(r)

	var req struct {
		PublicToken string `json:"public_token"`
	}
	if err := h.decodeJSON(r, &req); err != nil {
		h.sendPlaidError(w, http.StatusBadRequest, "INVALID_INPUT", "INVALID_REQUEST_BODY", err.Error())
		return
	}

	item, err := h.store.GetItemByPublicToken(r.Context(), req.PublicToken)
	if err == storage.ErrItemNotFound {
		h.sendPlaidError(w, http.StatusBadRequest, "INVALID_INPUT", "INVALID_PUBLIC_TOKEN", "unknown public_token")
		return
	}
	if err != nil {
		h.sendPlaidError(w, http.StatusInternalServerError, "INTERNAL", "LOOKUP_FAILED", err.Error())
		return
	}

	h.sendJSON(w, http.StatusOK, map[string]interface{}{
		"access_token": item.AccessToken,
		"item_id":      item.ItemID,
		"request_id":   requestID(),
	})
}

// ItemGet handles POST /item/get. Returns the item's institution id.
func (h *Handler) ItemGet(w http.ResponseWriter, r *http.Request) {
	h.logCreds(r)

	var req accessTokenRequest
	if err := h.decodeJSON(r, &req); err != nil {
		h.sendPlaidError(w, http.StatusBadRequest, "INVALID_INPUT", "INVALID_REQUEST_BODY", err.Error())
		return
	}

	item, err := h.requireItem(w, r, req.AccessToken)
	if err != nil {
		return
	}

	h.sendJSON(w, http.StatusOK, map[string]interface{}{
		"item": map[string]interface{}{
			"item_id":        item.ItemID,
			"institution_id": item.InstitutionID,
		},
		"request_id": requestID(),
	})
}

// InstitutionsGetByID handles POST /institutions/get_by_id. Returns the mock
// institution display name.
func (h *Handler) InstitutionsGetByID(w http.ResponseWriter, r *http.Request) {
	h.logCreds(r)

	var req struct {
		InstitutionID string   `json:"institution_id"`
		CountryCodes  []string `json:"country_codes"`
	}
	if err := h.decodeJSON(r, &req); err != nil {
		h.sendPlaidError(w, http.StatusBadRequest, "INVALID_INPUT", "INVALID_REQUEST_BODY", err.Error())
		return
	}

	inst, ok := models.Catalog[req.InstitutionID]
	if !ok {
		h.sendPlaidError(w, http.StatusBadRequest, "INVALID_INPUT", "INSTITUTION_NOT_FOUND", "unknown institution_id")
		return
	}

	h.sendJSON(w, http.StatusOK, map[string]interface{}{
		"institution": map[string]interface{}{
			"institution_id": inst.ID,
			"name":           inst.Name,
		},
		"request_id": requestID(),
	})
}

// RemoveItem handles POST /item/remove. Idempotent: removing a missing item is a
// no-op success (matches Plaid + the backend's best-effort disconnect).
func (h *Handler) RemoveItem(w http.ResponseWriter, r *http.Request) {
	h.logCreds(r)

	var req accessTokenRequest
	if err := h.decodeJSON(r, &req); err != nil {
		h.sendPlaidError(w, http.StatusBadRequest, "INVALID_INPUT", "INVALID_REQUEST_BODY", err.Error())
		return
	}

	if err := h.store.DeleteItemByAccessToken(r.Context(), req.AccessToken); err != nil {
		h.sendPlaidError(w, http.StatusInternalServerError, "INTERNAL", "REMOVE_FAILED", err.Error())
		return
	}

	h.sendJSON(w, http.StatusOK, map[string]interface{}{
		"request_id": requestID(),
	})
}

// requireItem resolves an item by access token or writes a Plaid 400 and returns
// an error so callers can return early.
func (h *Handler) requireItem(w http.ResponseWriter, r *http.Request, accessToken string) (models.Item, error) {
	item, err := h.store.GetItemByAccessToken(r.Context(), accessToken)
	if err == storage.ErrItemNotFound {
		h.sendPlaidError(w, http.StatusBadRequest, "INVALID_INPUT", "INVALID_ACCESS_TOKEN", "unknown access_token")
		return models.Item{}, err
	}
	if err != nil {
		h.sendPlaidError(w, http.StatusInternalServerError, "INTERNAL", "LOOKUP_FAILED", err.Error())
		return models.Item{}, err
	}
	return item, nil
}
