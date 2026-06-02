package ops

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"gitlab.com/fynbos/backend/api/apperrors"
	"gitlab.com/fynbos/backend/errcodes"
	"gitlab.com/fynbos/backend/providers/plaid"
	"gitlab.com/fynbos/backend/user/ops"
	"gitlab.com/fynbos/log"
	"go.uber.org/zap"
)

var (
	errMintFailed  = errors.New("plaid processor token mint failed")
	errFiantFailed = errors.New("fiant registration failed")
)

type Handlers struct {
	client    plaid.Client
	store     plaid.TokenStore
	linker    FiantLinker
	processor string
}

func New(client plaid.Client, store plaid.TokenStore, linker FiantLinker, processor string) *Handlers {
	return &Handlers{client: client, store: store, linker: linker, processor: processor}
}

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

func (h *Handlers) onPlaidErr(w http.ResponseWriter, r *http.Request, endpoint, userID string, err error) {
	log.Error("plaid: SDK call failed",
		zap.String("endpoint", endpoint),
		zap.String("user_id", userID),
		zap.Error(err),
	)
	apperrors.WriteAppError(w, r, http.StatusBadGateway, errcodes.ErrCodeInternal, "plaid request failed")
}

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
// Registers a Plaid-authorised account with Fiant by
// minting a `processor/token` and forwarding it via Fiant's payment-information
// endpoint.Returns:
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

	_, accessToken, ok := h.requireLinkedUser(w, r)
	if !ok {
		return
	}

	var (
		result        *LinkedIDs
		alreadyLinked bool
	)
	err = h.linker.WithAccountLock(r.Context(), u.ID, body.AccountID, func(ctx context.Context) error {
		existing, err := h.linker.ExistingLink(ctx, u.ID, body.AccountID)
		if err != nil {
			return err
		}
		if existing != nil {
			result, alreadyLinked = existing, true
			return nil
		}

		processorToken, err := h.client.CreateProcessorToken(ctx, accessToken, body.AccountID, h.processor)
		if err != nil {
			return fmt.Errorf("%w: %w", errMintFailed, err)
		}

		ids, err := h.linker.Register(ctx, LinkPlaidArgs{
			UserID:         u.ID,
			PlaidAccountID: body.AccountID,
			AccountName:    body.AccountName,
			AccountMask:    body.AccountMask,
			ProcessorToken: processorToken,
		})
		if err != nil {
			return fmt.Errorf("%w: %w", errFiantFailed, err)
		}
		result = ids
		return nil
	})
	if err != nil {
		switch {
		case errors.Is(err, errMintFailed):
			log.Error("plaid: CreateProcessorToken failed",
				zap.String("user_id", u.ID),
				zap.String("plaid_account_id", body.AccountID),
				zap.String("processor", h.processor),
				zap.Error(err),
			)
			apperrors.WriteAppError(w, r, http.StatusBadGateway, errcodes.ErrCodeInternal, "plaid processor token mint failed")
		case errors.Is(err, errFiantFailed):
			log.Error("plaid: FiantLinker.Register failed",
				zap.String("user_id", u.ID),
				zap.String("plaid_account_id", body.AccountID),
				zap.Error(err),
			)
			apperrors.WriteAppError(w, r, http.StatusBadGateway, errcodes.ErrCodeInternal, "fiant registration failed")
		default:
			log.Error("plaid: link-to-fiant failed",
				zap.String("user_id", u.ID),
				zap.String("plaid_account_id", body.AccountID),
				zap.Error(err),
			)
			apperrors.WriteAppError(w, r, http.StatusInternalServerError, errcodes.ErrCodeInternal, "failed to link plaid account to fiant")
		}
		return
	}

	if alreadyLinked {
		log.Info("plaid: account already linked to fiant",
			zap.String("user_id", u.ID),
			zap.String("plaid_account_id", body.AccountID),
			zap.String("linked_account_id", result.LinkedAccountID),
		)
		writeJSON(w, http.StatusOK, linkToFiantResponse{LinkedIDs: *result, AlreadyLinked: true})
		return
	}

	log.Info("plaid: account linked to fiant",
		zap.String("user_id", u.ID),
		zap.String("plaid_account_id", body.AccountID),
		zap.String("linked_account_id", result.LinkedAccountID),
		zap.String("payment_information_id", result.PaymentInformationID),
	)
	writeJSON(w, http.StatusOK, linkToFiantResponse{LinkedIDs: *result, AlreadyLinked: false})
}

type linkToFiantResponse struct {
	LinkedIDs
	AlreadyLinked bool `json:"already_linked"`
}

// Returns the Plaid account_ids the current user has already linked
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

// Disconnect removes the Item on Plaid's side and always deletes the local
// TokenStore entry. Plaid's ItemRemove is soft-failed: it returns
// `{"disconnected": true}` once the local store is clean even if Plaid returned
// an error, so a partial failure can never leave a user permanently stuck.
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

	// TODO: inform Fiant about the token removal; this method may belong in the plaid-fiant client.

	log.Info("plaid item disconnected",
		zap.String("user_id", u.ID),
		zap.String("item_id", t.ItemID),
	)
	writeJSON(w, http.StatusOK, struct {
		Disconnected bool `json:"disconnected"`
	}{Disconnected: true})
}
