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
// Each method is filled in by tasks B5a–B5f; B4 mounts them. Phase 2 adds
// `linker` + `processor` for /plaid/link-to-fiant; both may be nil/empty when
// the Phase-2 wiring is absent (router conditionally registers the route).
type Handlers struct {
	client    plaid.Client
	store     plaid.TokenStore
	linker    FiantLinker
	processor string
}

// New wires a Handlers with its collaborators. `linker` and `processor` are
// optional and only required for the Phase-2 /plaid/link-to-fiant route.
func New(client plaid.Client, store plaid.TokenStore, linker FiantLinker, processor string) *Handlers {
	return &Handlers{client: client, store: store, linker: linker, processor: processor}
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

// GetTransactions — GET /plaid/transactions. Drives `/transactions/sync` from
// cursor=0 and returns the full added/modified/removed roll-up plus the final
// next_cursor. Sandbox needs a few seconds after item creation to populate;
// an empty array with a cursor is the normal early response.
func (h *Handlers) GetTransactions(w http.ResponseWriter, r *http.Request) {
	userID, accessToken, ok := h.requireLinkedUser(w, r)
	if !ok {
		return
	}
	res, err := h.client.SyncTransactions(r.Context(), accessToken)
	if err != nil {
		h.onPlaidErr(w, r, "TransactionsSync", userID, err)
		return
	}
	writeJSON(w, http.StatusOK, res)
}

// LinkToFiant — POST /plaid/link-to-fiant. Body:
//
//	{ "account_id": "...", "account_name": "...", "account_mask": "..." }
//
// Phase 2 of the Plaid POC. Registers a Plaid-authorised account with Fiant by
// minting a `processor/token` and forwarding it via Fiant's payment-information
// endpoint. Idempotent on `(wallet_id, plaid_account_id)` — see B13's partial
// unique index. Returns:
//
//	{ "linked_account_id": "...", "payment_information_id": "...", "already_linked": bool }
//
// `already_linked: true` means the row existed; no Plaid or Fiant calls were
// made on this request.
func (h *Handlers) LinkToFiant(w http.ResponseWriter, r *http.Request) {
	if h.linker == nil || h.processor == "" {
		apperrors.WriteAppError(w, r, http.StatusServiceUnavailable, errcodes.ErrCodeInternal, "plaid/fiant linker not configured")
		return
	}

	u, err := ops.UserForContext(r.Context())
	if err != nil {
		apperrors.WriteAppError(w, r, http.StatusUnauthorized, errcodes.ErrCodeUnauthorized, "unauthenticated")
		return
	}

	var body struct {
		AccountID   string `json:"account_id"`
		AccountName string `json:"account_name"`
		AccountMask string `json:"account_mask"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.AccountID == "" {
		apperrors.WriteAppError(w, r, http.StatusBadRequest, errcodes.ErrCodeBadRequest, "account_id is required")
		return
	}

	// (a) Dedupe first — cheapest path, no Plaid/Fiant network calls.
	existing, err := h.linker.ExistingLink(r.Context(), u.ID, body.AccountID)
	if err != nil {
		log.Error("plaid: ExistingLink lookup failed",
			zap.String("user_id", u.ID),
			zap.String("plaid_account_id", body.AccountID),
			zap.Error(err),
		)
		apperrors.WriteAppError(w, r, http.StatusInternalServerError, errcodes.ErrCodeInternal, "failed to check existing fiant link")
		return
	}
	if existing != nil {
		log.Info("plaid: account already linked to fiant (idempotent hit)",
			zap.String("user_id", u.ID),
			zap.String("plaid_account_id", body.AccountID),
			zap.String("linked_account_id", existing.LinkedAccountID),
		)
		writeJSON(w, http.StatusOK, linkToFiantResponse{LinkedIDs: *existing, AlreadyLinked: true})
		return
	}

	// (b) Need the access_token. 404 cleanly if the user hasn't done Phase 1.
	_, accessToken, ok := h.requireLinkedUser(w, r)
	if !ok {
		return
	}

	// (c) Mint a processor token. Long-lived, bound to (item, account, processor).
	processorToken, err := h.client.CreateProcessorToken(r.Context(), accessToken, body.AccountID, h.processor)
	if err != nil {
		log.Error("plaid: CreateProcessorToken failed",
			zap.String("user_id", u.ID),
			zap.String("plaid_account_id", body.AccountID),
			zap.String("processor", h.processor),
			zap.Error(err),
		)
		apperrors.WriteAppError(w, r, http.StatusBadGateway, errcodes.ErrCodeInternal, "plaid processor token mint failed")
		return
	}

	// (d) Cross-system write: Fiant POST + linked_account persist. Idempotency
	//     guard at the DB layer via the partial unique index — duplicate inserts
	//     surface as ErrInternal here but should be rare given (a).
	ids, err := h.linker.Register(r.Context(), LinkPlaidArgs{
		UserID:         u.ID,
		PlaidAccountID: body.AccountID,
		AccountName:    body.AccountName,
		AccountMask:    body.AccountMask,
		ProcessorToken: processorToken,
	})
	if err != nil {
		log.Error("plaid: FiantLinker.Register failed",
			zap.String("user_id", u.ID),
			zap.String("plaid_account_id", body.AccountID),
			zap.Error(err),
		)
		apperrors.WriteAppError(w, r, http.StatusBadGateway, errcodes.ErrCodeInternal, "fiant registration failed")
		return
	}

	log.Info("plaid: account linked to fiant",
		zap.String("user_id", u.ID),
		zap.String("plaid_account_id", body.AccountID),
		zap.String("linked_account_id", ids.LinkedAccountID),
		zap.String("payment_information_id", ids.PaymentInformationID),
	)
	writeJSON(w, http.StatusOK, linkToFiantResponse{LinkedIDs: *ids, AlreadyLinked: false})
}

type linkToFiantResponse struct {
	LinkedIDs
	AlreadyLinked bool `json:"already_linked"`
}

// GetRegistered — GET /plaid/registered. Returns the Plaid account_ids the
// current user has already linked to Fiant via /plaid/link-to-fiant. Drives
// the "Linked" tag on the /connect/plaid/{country} selector. Always returns
// 200 with a (possibly empty) array; never reveals counts of other users.
func (h *Handlers) GetRegistered(w http.ResponseWriter, r *http.Request) {
	if h.linker == nil {
		apperrors.WriteAppError(w, r, http.StatusServiceUnavailable, errcodes.ErrCodeInternal, "plaid/fiant linker not configured")
		return
	}

	u, err := ops.UserForContext(r.Context())
	if err != nil {
		apperrors.WriteAppError(w, r, http.StatusUnauthorized, errcodes.ErrCodeUnauthorized, "unauthenticated")
		return
	}

	ids, err := h.linker.ListLinkedPlaidAccountIDs(r.Context(), u.ID)
	if err != nil {
		log.Error("plaid: ListLinkedPlaidAccountIDs failed",
			zap.String("user_id", u.ID),
			zap.Error(err),
		)
		apperrors.WriteAppError(w, r, http.StatusInternalServerError, errcodes.ErrCodeInternal, "failed to list registered plaid accounts")
		return
	}
	if ids == nil {
		ids = []string{}
	}
	writeJSON(w, http.StatusOK, struct {
		PlaidAccountIDs []string `json:"plaid_account_ids"`
	}{PlaidAccountIDs: ids})
}

// Disconnect — DELETE /plaid/disconnect. Removes the Item on Plaid's side
// (best-effort) and always deletes the local TokenStore entry. Returns
// `{"disconnected": true}` once the store is clean even if Plaid returned an
// error, so a partial failure can never leave a user permanently stuck.
func (h *Handlers) Disconnect(w http.ResponseWriter, r *http.Request) {
	u, err := ops.UserForContext(r.Context())
	if err != nil {
		apperrors.WriteAppError(w, r, http.StatusUnauthorized, errcodes.ErrCodeUnauthorized, "unauthenticated")
		return
	}

	t, found, err := h.store.Get(r.Context(), u.ID)
	if err != nil {
		log.Error("plaid: TokenStore.Get failed",
			zap.String("user_id", u.ID),
			zap.Error(err),
		)
		apperrors.WriteAppError(w, r, http.StatusInternalServerError, errcodes.ErrCodeInternal, "failed to read plaid state")
		return
	}

	if found {
		if err := h.client.RemoveItem(r.Context(), t.AccessToken); err != nil {
			// Soft-fail: token may already be invalid on Plaid's side; we still
			// want to drop our local record so the user can re-link.
			log.Warn("plaid: ItemRemove failed (continuing with local delete)",
				zap.String("user_id", u.ID),
				zap.String("item_id", t.ItemID),
				zap.Error(err),
			)
		}
	}

	if err := h.store.Delete(r.Context(), u.ID); err != nil {
		log.Error("plaid: TokenStore.Delete failed",
			zap.String("user_id", u.ID),
			zap.Error(err),
		)
		apperrors.WriteAppError(w, r, http.StatusInternalServerError, errcodes.ErrCodeInternal, "failed to clear plaid link")
		return
	}

	log.Info("plaid item disconnected",
		zap.String("user_id", u.ID),
		zap.String("item_id", t.ItemID),
	)
	writeJSON(w, http.StatusOK, struct {
		Disconnected bool `json:"disconnected"`
	}{Disconnected: true})
}
