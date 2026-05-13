// Package ops contains the HTTP handler layer for the Plaid POC provider.
// Handlers are thin: they extract the user from request context, delegate to
// plaid.Client + plaid.TokenStore, and marshal the response. No SDK or
// persistence concerns belong here.
package ops

import (
	"encoding/json"
	"net/http"
	"time"

	"gitlab.com/fynbos/backend/api/apperrors"
	"gitlab.com/fynbos/backend/errcodes"
	"gitlab.com/fynbos/backend/providers/plaid"
	"gitlab.com/fynbos/backend/user/ops"
	"gitlab.com/fynbos/log"
	"go.uber.org/zap"
)

// Handlers groups the chi handler funcs that make up the /plaid HTTP surface.
// Each method is filled in by tasks B5a–B5f; B4 mounts them.
type Handlers struct {
	client plaid.Client
	store  plaid.TokenStore
}

// New wires a Handlers with its collaborators.
func New(client plaid.Client, store plaid.TokenStore) *Handlers {
	return &Handlers{client: client, store: store}
}

// notImplemented is used by every endpoint until its B5* task fills it in.
// Logs a debug line so we can see the call without leaking tokens.
func (h *Handlers) notImplemented(w http.ResponseWriter, r *http.Request, endpoint string) {
	u, err := ops.UserForContext(r.Context())
	if err != nil {
		apperrors.WriteAppError(w, r, http.StatusUnauthorized, errcodes.ErrCodeUnauthorized, "unauthenticated")
		return
	}
	log.Debug("plaid endpoint stub hit",
		zap.String("endpoint", endpoint),
		zap.String("user_id", u.ID),
	)
	apperrors.WriteAppError(w, r, http.StatusNotImplemented, errcodes.ErrCodeInternal, "plaid endpoint not yet implemented")
}

// CreateLinkToken — POST /plaid/link-token. Calls Plaid /link/token/create
// using the current Kratos user as client_user_id and returns the link token
// the frontend hands to react-plaid-link.
func (h *Handlers) CreateLinkToken(w http.ResponseWriter, r *http.Request) {
	u, err := ops.UserForContext(r.Context())
	if err != nil {
		apperrors.WriteAppError(w, r, http.StatusUnauthorized, errcodes.ErrCodeUnauthorized, "unauthenticated")
		return
	}

	linkToken, expiration, err := h.client.CreateLinkToken(r.Context(), u.ID)
	if err != nil {
		log.Error("plaid: CreateLinkToken failed",
			zap.String("user_id", u.ID),
			zap.Error(err),
		)
		apperrors.WriteAppError(w, r, http.StatusBadGateway, errcodes.ErrCodeInternal, "plaid link-token create failed")
		return
	}

	log.Info("plaid link token issued",
		zap.String("user_id", u.ID),
		zap.Time("expiration", expiration),
	)
	writeJSON(w, http.StatusOK, struct {
		LinkToken  string    `json:"link_token"`
		Expiration time.Time `json:"expiration"`
	}{LinkToken: linkToken, Expiration: expiration})
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(body); err != nil {
		log.Error("plaid: writeJSON encode", zap.Error(err))
	}
}

// ExchangePublicToken — POST /plaid/exchange. Body: {"public_token": "..."}.
// Exchanges with Plaid, fetches institution metadata, stores TokenSet keyed by
// the current Kratos user. Never leaks tokens to logs.
func (h *Handlers) ExchangePublicToken(w http.ResponseWriter, r *http.Request) {
	u, err := ops.UserForContext(r.Context())
	if err != nil {
		apperrors.WriteAppError(w, r, http.StatusUnauthorized, errcodes.ErrCodeUnauthorized, "unauthenticated")
		return
	}

	var body struct {
		PublicToken string `json:"public_token"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.PublicToken == "" {
		apperrors.WriteAppError(w, r, http.StatusBadRequest, errcodes.ErrCodeBadRequest, "public_token is required")
		return
	}

	accessToken, itemID, err := h.client.ExchangePublicToken(r.Context(), body.PublicToken)
	if err != nil {
		log.Error("plaid: ExchangePublicToken failed",
			zap.String("user_id", u.ID),
			zap.Error(err),
		)
		apperrors.WriteAppError(w, r, http.StatusBadGateway, errcodes.ErrCodeInternal, "plaid public-token exchange failed")
		return
	}

	institutionID, institutionName, err := h.client.GetInstitutionForItem(r.Context(), accessToken)
	if err != nil {
		// Soft-fail: token exchange succeeded; institution lookup is decoration.
		log.Warn("plaid: GetInstitutionForItem failed (continuing)",
			zap.String("user_id", u.ID),
			zap.String("item_id", itemID),
			zap.Error(err),
		)
	}

	h.store.Put(u.ID, plaid.TokenSet{
		AccessToken:     accessToken,
		ItemID:          itemID,
		InstitutionID:   institutionID,
		InstitutionName: institutionName,
		LinkedAt:        time.Now().UTC(),
	})

	log.Info("plaid item linked",
		zap.String("user_id", u.ID),
		zap.String("item_id", itemID),
		zap.String("institution_id", institutionID),
		zap.String("institution_name", institutionName),
	)
	writeJSON(w, http.StatusOK, struct {
		ItemID          string `json:"item_id"`
		InstitutionName string `json:"institution_name"`
	}{ItemID: itemID, InstitutionName: institutionName})
}

// GetState — GET /plaid/state. Returns a non-sensitive view of the current
// user's Plaid link: linked? what institution? when? Never includes tokens.
func (h *Handlers) GetState(w http.ResponseWriter, r *http.Request) {
	u, err := ops.UserForContext(r.Context())
	if err != nil {
		apperrors.WriteAppError(w, r, http.StatusUnauthorized, errcodes.ErrCodeUnauthorized, "unauthenticated")
		return
	}

	t, ok := h.store.Get(u.ID)
	if !ok {
		writeJSON(w, http.StatusOK, plaid.State{Linked: false})
		return
	}
	linkedAt := t.LinkedAt
	writeJSON(w, http.StatusOK, plaid.State{
		Linked:          true,
		ItemID:          t.ItemID,
		InstitutionName: t.InstitutionName,
		LinkedAt:        &linkedAt,
	})
}

// GetAccounts — filled by B5d.
func (h *Handlers) GetAccounts(w http.ResponseWriter, r *http.Request) {
	h.notImplemented(w, r, "GET /plaid/accounts")
}

// GetAuth — filled by B5d.
func (h *Handlers) GetAuth(w http.ResponseWriter, r *http.Request) {
	h.notImplemented(w, r, "GET /plaid/auth")
}

// GetBalance — filled by B5d.
func (h *Handlers) GetBalance(w http.ResponseWriter, r *http.Request) {
	h.notImplemented(w, r, "GET /plaid/balance")
}

// GetIdentity — filled by B5d.
func (h *Handlers) GetIdentity(w http.ResponseWriter, r *http.Request) {
	h.notImplemented(w, r, "GET /plaid/identity")
}

// GetTransactions — filled by B5e.
func (h *Handlers) GetTransactions(w http.ResponseWriter, r *http.Request) {
	h.notImplemented(w, r, "GET /plaid/transactions")
}

// Disconnect — filled by B5f.
func (h *Handlers) Disconnect(w http.ResponseWriter, r *http.Request) {
	h.notImplemented(w, r, "DELETE /plaid/disconnect")
}
