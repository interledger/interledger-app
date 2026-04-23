package v1

import (
	"io"
	"net/http"

	"gitlab.com/fynbos/backend/api/apperrors"
	"gitlab.com/fynbos/log"
	"go.uber.org/zap"
)

func (h *handlers) getAccountConfirmation(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if _, err := h.users.UserForContext(ctx); err != nil {
		apperrors.ToHTTPError(w, r, err)
		return
	}

	wallet, err := h.wallets.ForContext(ctx)
	if err != nil {
		apperrors.ToHTTPError(w, r, err)
		return
	}

	body, err := h.gatehub.GetAccountConfirmation(ctx, wallet.ID)
	if err != nil {
		apperrors.ToHTTPError(w, r, err)
		return
	}
	defer body.Close()

	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", "inline; filename=\"account-confirmation.pdf\"")
	w.WriteHeader(http.StatusOK)

	if _, err := io.Copy(w, body); err != nil {
		log.Warn("failed to stream account confirmation", zap.Error(err))
	}
}
