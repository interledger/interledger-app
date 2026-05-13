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

	if err := h.store.Put(r.Context(), u.ID, plaid.TokenSet{
		AccessToken:     accessToken,
		ItemID:          itemID,
		InstitutionID:   institutionID,
		InstitutionName: institutionName,
		LinkedAt:        time.Now().UTC(),
	}); err != nil {
		log.Error("plaid: TokenStore.Put failed",
			zap.String("user_id", u.ID),
			zap.String("item_id", itemID),
			zap.Error(err),
		)
		apperrors.WriteAppError(w, r, http.StatusInternalServerError, errcodes.ErrCodeInternal, "failed to persist plaid link")
		return
	}

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

	t, ok, err := h.store.Get(r.Context(), u.ID)
	if err != nil {
		log.Error("plaid: TokenStore.Get failed",
			zap.String("user_id", u.ID),
			zap.Error(err),
		)
		apperrors.WriteAppError(w, r, http.StatusInternalServerError, errcodes.ErrCodeInternal, "failed to read plaid state")
		return
	}
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

// requireLinkedUser resolves the Kratos user and their stored TokenSet.
// Writes the appropriate HTTP error and returns ok=false if either is missing
// or the store call failed; caller short-circuits on ok=false.
func (h *Handlers) requireLinkedUser(w http.ResponseWriter, r *http.Request) (userID, accessToken string, ok bool) {
	u, err := ops.UserForContext(r.Context())
	if err != nil {
		apperrors.WriteAppError(w, r, http.StatusUnauthorized, errcodes.ErrCodeUnauthorized, "unauthenticated")
		return "", "", false
	}
	t, found, err := h.store.Get(r.Context(), u.ID)
	if err != nil {
		log.Error("plaid: TokenStore.Get failed",
			zap.String("user_id", u.ID),
			zap.Error(err),
		)
		apperrors.WriteAppError(w, r, http.StatusInternalServerError, errcodes.ErrCodeInternal, "failed to read plaid state")
		return "", "", false
	}
	if !found {
		apperrors.WriteAppError(w, r, http.StatusNotFound, errcodes.ErrCodeNotFound, "no plaid item linked for this user")
		return "", "", false
	}
	return u.ID, t.AccessToken, true
}

// onPlaidErr logs and writes a 502 with no token detail.
func (h *Handlers) onPlaidErr(w http.ResponseWriter, r *http.Request, endpoint, userID string, err error) {
	log.Error("plaid: SDK call failed",
		zap.String("endpoint", endpoint),
		zap.String("user_id", userID),
		zap.Error(err),
	)
	apperrors.WriteAppError(w, r, http.StatusBadGateway, errcodes.ErrCodeInternal, "plaid request failed")
}

// GetAccounts — GET /plaid/accounts. Returns Plaid `/accounts/get` verbatim.
func (h *Handlers) GetAccounts(w http.ResponseWriter, r *http.Request) {
	userID, accessToken, ok := h.requireLinkedUser(w, r)
	if !ok {
		return
	}
	resp, err := h.client.GetAccounts(r.Context(), accessToken)
	if err != nil {
		h.onPlaidErr(w, r, "AccountsGet", userID, err)
		return
	}
	writeJSON(w, http.StatusOK, resp)
}

// GetAuth — GET /plaid/auth. Returns Plaid `/auth/get` verbatim. Includes ACH
// routing + account numbers — sensitive but Plaid's own surface, so we don't
// add extra masking.
func (h *Handlers) GetAuth(w http.ResponseWriter, r *http.Request) {
	userID, accessToken, ok := h.requireLinkedUser(w, r)
	if !ok {
		return
	}
	resp, err := h.client.GetAuth(r.Context(), accessToken)
	if err != nil {
		h.onPlaidErr(w, r, "AuthGet", userID, err)
		return
	}
	writeJSON(w, http.StatusOK, resp)
}

// GetBalance — GET /plaid/balance. Returns Plaid `/accounts/balance/get`
// verbatim (forces a fresh balance refresh).
func (h *Handlers) GetBalance(w http.ResponseWriter, r *http.Request) {
	userID, accessToken, ok := h.requireLinkedUser(w, r)
	if !ok {
		return
	}
	resp, err := h.client.GetBalance(r.Context(), accessToken)
	if err != nil {
		h.onPlaidErr(w, r, "AccountsBalanceGet", userID, err)
		return
	}
	writeJSON(w, http.StatusOK, resp)
}

// GetIdentity — GET /plaid/identity. Returns Plaid `/identity/get` verbatim.
func (h *Handlers) GetIdentity(w http.ResponseWriter, r *http.Request) {
	userID, accessToken, ok := h.requireLinkedUser(w, r)
	if !ok {
		return
	}
	resp, err := h.client.GetIdentity(r.Context(), accessToken)
	if err != nil {
		h.onPlaidErr(w, r, "IdentityGet", userID, err)
		return
	}
	writeJSON(w, http.StatusOK, resp)
}

// GetTransactions — filled by B5e.
func (h *Handlers) GetTransactions(w http.ResponseWriter, r *http.Request) {
	h.notImplemented(w, r, "GET /plaid/transactions")
}

// Disconnect — filled by B5f.
func (h *Handlers) Disconnect(w http.ResponseWriter, r *http.Request) {
	h.notImplemented(w, r, "DELETE /plaid/disconnect")
}
