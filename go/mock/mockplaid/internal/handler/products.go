package handler

import (
	"net/http"

	"gitlab.com/fynbos/mock/mockplaid/internal/models"
)

// accountObj serialises one stored account into the Plaid AccountBase shape
// (with static balances). account_id/name/mask/type/subtype come from the item.
func accountObj(a models.Account) map[string]interface{} {
	return map[string]interface{}{
		"account_id": a.AccountID,
		"name":       a.Name,
		"mask":       a.Mask,
		"type":       a.Type,
		"subtype":    a.Subtype,
		"balances": map[string]interface{}{
			"available":                100.0,
			"current":                  110.0,
			"iso_currency_code":        "USD",
			"limit":                    nil,
			"unofficial_currency_code": nil,
		},
	}
}

func accountsArray(item models.Item) []map[string]interface{} {
	out := make([]map[string]interface{}, 0, len(item.Accounts))
	for _, a := range item.Accounts {
		out = append(out, accountObj(a))
	}
	return out
}

func itemObj(item models.Item) map[string]interface{} {
	return map[string]interface{}{
		"item_id":        item.ItemID,
		"institution_id": item.InstitutionID,
	}
}

// GetAccounts handles POST /accounts/get.
func (h *Handler) GetAccounts(w http.ResponseWriter, r *http.Request) {
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
		"accounts":   accountsArray(item),
		"item":       itemObj(item),
		"request_id": requestID(),
	})
}

// GetBalance handles POST /accounts/balance/get (same shape as accounts/get).
func (h *Handler) GetBalance(w http.ResponseWriter, r *http.Request) {
	h.GetAccounts(w, r)
}

// GetAuth handles POST /auth/get — accounts + ACH numbers.
func (h *Handler) GetAuth(w http.ResponseWriter, r *http.Request) {
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
	ach := make([]map[string]interface{}, 0, len(item.Accounts))
	for _, a := range item.Accounts {
		ach = append(ach, map[string]interface{}{
			"account_id":   a.AccountID,
			"account":      "1111222233330000",
			"routing":      "011401533",
			"wire_routing": "021000021",
		})
	}
	h.sendJSON(w, http.StatusOK, map[string]interface{}{
		"accounts": accountsArray(item),
		"numbers": map[string]interface{}{
			"ach":           ach,
			"eft":           []interface{}{},
			"international": []interface{}{},
			"bacs":          []interface{}{},
		},
		"item":       itemObj(item),
		"request_id": requestID(),
	})
}

// GetIdentity handles POST /identity/get — accounts with owners.
func (h *Handler) GetIdentity(w http.ResponseWriter, r *http.Request) {
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
	owners := []map[string]interface{}{
		{
			"names":         []string{"Alex Mock"},
			"phone_numbers": []interface{}{},
			"emails": []map[string]interface{}{
				{"data": "alex@example.test", "primary": true, "type": "primary"},
			},
			"addresses": []interface{}{},
		},
	}
	accounts := accountsArray(item)
	for i := range accounts {
		accounts[i]["owners"] = owners
	}
	h.sendJSON(w, http.StatusOK, map[string]interface{}{
		"accounts":   accounts,
		"item":       itemObj(item),
		"request_id": requestID(),
	})
}

// TransactionsSync handles POST /transactions/sync — a single terminal page.
func (h *Handler) TransactionsSync(w http.ResponseWriter, r *http.Request) {
	h.logCreds(r)
	var req struct {
		AccessToken string `json:"access_token"`
		Cursor      string `json:"cursor"`
	}
	if err := h.decodeJSON(r, &req); err != nil {
		h.sendPlaidError(w, http.StatusBadRequest, "INVALID_INPUT", "INVALID_REQUEST_BODY", err.Error())
		return
	}
	item, err := h.requireItem(w, r, req.AccessToken)
	if err != nil {
		return
	}

	added := []interface{}{}
	// On the first sync (no cursor) return one static transaction; re-syncs are empty.
	if req.Cursor == "" && len(item.Accounts) > 0 {
		added = append(added, map[string]interface{}{
			"transaction_id":    "txn_mock_1",
			"account_id":        item.Accounts[0].AccountID,
			"amount":            12.34,
			"iso_currency_code": "USD",
			"name":              "Mock Coffee",
			"date":              "2026-06-01",
			"pending":           false,
		})
	}

	h.sendJSON(w, http.StatusOK, map[string]interface{}{
		"added":       added,
		"modified":    []interface{}{},
		"removed":     []interface{}{},
		"next_cursor": "mockplaid-cursor-end",
		"has_more":    false,
		"request_id":  requestID(),
	})
}
