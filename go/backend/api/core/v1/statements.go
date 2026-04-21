package v1

import (
	"io"
	"net/http"

	api_middleware "gitlab.com/fynbos/backend/api/middleware"
)

func (h *handlers) getAccountStatement(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	_, err := h.users.UserForContext(ctx)
	if err != nil {
		api_middleware.WriteAppError(w, r, http.StatusUnauthorized, api_middleware.ErrCodeUnauthorized, "unauthorized")
		return
	}

	wallet, err := h.wallets.ForContext(ctx)
	if err != nil {
		api_middleware.WriteAppError(w, r, http.StatusUnauthorized, api_middleware.ErrCodeUnauthorized, "unauthorized")
		return
	}

	body, err := h.gatehub.GetAccountStatement(ctx, wallet.ID)
	if err != nil {
		toHTTPError(w, r, err)
		return
	}
	defer body.Close()

	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", "attachment; filename=\"account-statement.pdf\"")
	w.WriteHeader(http.StatusOK)

	_, _ = io.Copy(w, body)
}
