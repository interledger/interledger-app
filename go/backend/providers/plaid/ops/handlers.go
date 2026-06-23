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
	errExchangeFailed = errors.New("plaid public-token exchange failed")
	errMintFailed     = errors.New("plaid processor token mint failed")
	errFiantFailed    = errors.New("fiant registration failed")
)

type Handlers struct {
	client    plaid.Client
	linker    FiantLinker
	processor string
}

func New(client plaid.Client, linker FiantLinker, processor string) *Handlers {
	return &Handlers{client: client, linker: linker, processor: processor}
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

// LinkToFiant — POST /plaid/link-to-fiant. Body:
//
//	{ "public_token": "...", "account_id": "...", "account_name": "...", "account_mask": "..." }
//
// Exchanges the public_token in-request (no store), mints a processor token, and
// registers the account with Fiant. Returns:
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
		PublicToken string `json:"public_token"`
		AccountID   string `json:"account_id"`
		AccountName string `json:"account_name"`
		AccountMask string `json:"account_mask"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.AccountID == "" || body.PublicToken == "" {
		apperrors.WriteAppError(w, r, http.StatusBadRequest, errcodes.ErrCodeBadRequest, "public_token and account_id are required")
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

		accessToken, _, err := h.client.ExchangePublicToken(ctx, body.PublicToken)
		if err != nil {
			return fmt.Errorf("%w: %w", errExchangeFailed, err)
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
		case errors.Is(err, errExchangeFailed):
			log.Error("plaid: ExchangePublicToken failed in link-to-fiant",
				zap.String("user_id", u.ID),
				zap.String("plaid_account_id", body.AccountID),
				zap.Error(err),
			)
			apperrors.WriteAppError(w, r, http.StatusBadGateway, errcodes.ErrCodeInternal, "plaid public-token exchange failed")
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

