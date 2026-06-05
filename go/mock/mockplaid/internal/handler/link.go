package handler

import (
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"gitlab.com/fynbos/mock/mockplaid/internal/logger"
	"gitlab.com/fynbos/mock/mockplaid/internal/models"
)

type linkTokenCreateRequest struct {
	User struct {
		ClientUserID string `json:"client_user_id"`
	} `json:"user"`
	Products []string `json:"products"`
}

// CreateLinkToken handles POST /link/token/create. Mints a mock link token and
// records a link session the mock Link UI later resolves.
func (h *Handler) CreateLinkToken(w http.ResponseWriter, r *http.Request) {
	h.logCreds(r)

	var req linkTokenCreateRequest
	if err := h.decodeJSON(r, &req); err != nil {
		h.sendPlaidError(w, http.StatusBadRequest, "INVALID_INPUT", "INVALID_REQUEST_BODY", err.Error())
		return
	}

	linkToken := "link-sandbox-" + uuid.NewString()
	session := models.LinkSession{
		LinkToken: linkToken,
		UserID:    req.User.ClientUserID,
		CreatedAt: time.Now().UTC(),
	}
	if err := h.store.SaveLinkSession(r.Context(), session); err != nil {
		h.sendPlaidError(w, http.StatusInternalServerError, "INTERNAL", "SAVE_FAILED", err.Error())
		return
	}

	logger.Info("plaid link session created",
		zap.String("link_token", linkToken),
		zap.String("user_id", req.User.ClientUserID),
	)

	h.sendJSON(w, http.StatusOK, map[string]interface{}{
		"link_token": linkToken,
		"expiration": time.Now().UTC().Add(4 * time.Hour).Format(time.RFC3339),
		"request_id": requestID(),
	})
}

type selectRequest struct {
	LinkToken     string `json:"link_token"`
	InstitutionID string `json:"institution_id"`
	AccountKey    string `json:"account_key"`
}

// SelectAccount handles POST /link/session/select — a mock-only control endpoint
// the Link UI calls when the user picks a bank + account. Mints a public token and
// stores a fully-formed item. Bank A account IDs are stable per account; Bank B
// mints a fresh account ID every call.
func (h *Handler) SelectAccount(w http.ResponseWriter, r *http.Request) {
	var req selectRequest
	if err := h.decodeJSON(r, &req); err != nil {
		h.sendPlaidError(w, http.StatusBadRequest, "INVALID_INPUT", "INVALID_REQUEST_BODY", err.Error())
		return
	}

	if _, err := h.store.GetLinkSession(r.Context(), req.LinkToken); err != nil {
		h.sendPlaidError(w, http.StatusBadRequest, "INVALID_INPUT", "INVALID_LINK_TOKEN", "unknown or expired link_token")
		return
	}

	inst, ok := models.Catalog[req.InstitutionID]
	if !ok {
		h.sendPlaidError(w, http.StatusBadRequest, "INVALID_INPUT", "INVALID_INSTITUTION", "unknown institution_id")
		return
	}
	cat, ok := inst.Account(req.AccountKey)
	if !ok {
		h.sendPlaidError(w, http.StatusBadRequest, "INVALID_INPUT", "INVALID_ACCOUNT", "unknown account_key")
		return
	}

	// Determinism rule: Bank A is stable per account; Bank B is always-new.
	var accountID string
	if req.InstitutionID == models.InstitutionB {
		seq, err := h.store.NextAccountSeq(r.Context())
		if err != nil {
			h.sendPlaidError(w, http.StatusInternalServerError, "INTERNAL", "SEQ_FAILED", err.Error())
			return
		}
		accountID = fmt.Sprintf("acc_mock_b_%d", seq)
	} else {
		accountID = "acc_mock_a_" + req.AccountKey
	}

	acct := models.Account{
		AccountID: accountID,
		Name:      cat.Name,
		Mask:      cat.Mask,
		Type:      "depository",
		Subtype:   cat.Subtype,
	}

	publicToken := "public-sandbox-" + uuid.NewString()
	item := models.Item{
		AccessToken:     "access-sandbox-" + uuid.NewString(),
		ItemID:          "item-" + uuid.NewString(),
		InstitutionID:   inst.ID,
		InstitutionName: inst.Name,
		Accounts:        []models.Account{acct},
		PublicToken:     publicToken,
	}
	if err := h.store.SaveItem(r.Context(), item); err != nil {
		h.sendPlaidError(w, http.StatusInternalServerError, "INTERNAL", "SAVE_FAILED", err.Error())
		return
	}

	logger.Info("plaid link selection",
		zap.String("institution_id", inst.ID),
		zap.String("account_id", accountID),
		zap.String("item_id", item.ItemID),
	)

	h.sendJSON(w, http.StatusOK, map[string]interface{}{
		"public_token": publicToken,
		"metadata": map[string]interface{}{
			"institution": map[string]interface{}{
				"name":           inst.Name,
				"institution_id": inst.ID,
			},
			"accounts": []map[string]interface{}{
				{
					"id":      acct.AccountID,
					"name":    acct.Name,
					"mask":    acct.Mask,
					"type":    acct.Type,
					"subtype": acct.Subtype,
				},
			},
			"link_session_id": "link-session-" + uuid.NewString(),
		},
	})
}
