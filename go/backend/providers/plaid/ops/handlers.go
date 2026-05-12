// Package ops contains the HTTP handler layer for the Plaid POC provider.
// Handlers are thin: they extract the user from request context, delegate to
// plaid.Client + plaid.TokenStore, and marshal the response. No SDK or
// persistence concerns belong here.
package ops

import (
	"net/http"

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

// CreateLinkToken — filled by B5a.
func (h *Handlers) CreateLinkToken(w http.ResponseWriter, r *http.Request) {
	h.notImplemented(w, r, "POST /plaid/link-token")
}

// ExchangePublicToken — filled by B5b.
func (h *Handlers) ExchangePublicToken(w http.ResponseWriter, r *http.Request) {
	h.notImplemented(w, r, "POST /plaid/exchange")
}

// GetState — filled by B5c.
func (h *Handlers) GetState(w http.ResponseWriter, r *http.Request) {
	h.notImplemented(w, r, "GET /plaid/state")
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
